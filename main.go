package main

import (
	"context"
	goflag "flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"time"

	"vm-crd-create/controller"
	"vm-crd-create/pkg/driver"
	cloudnative "vm-crd-create/pkg/generated/clientset/versioned"
	cloudnativeinformer "vm-crd-create/pkg/generated/informers/externalversions"
	"vm-crd-create/pkg/version"

	flag "github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"
)

const defaultSyncTime = time.Second * 30

var (
	kubeconfig         string
	showVersion        bool
	threads            int
	leaderElect        bool
	id                 string
	leaseLockName      string
	leaseLockNamespace string
	driverName         string
)

func parseFlags() {
	flag.StringVarP(&kubeconfig, "kubeconfig", "c", "", "absolute path to the kubeconfig file")
	flag.IntVarP(&threads, "threads", "t", 2, "number of worker threads used by the controller")
	flag.StringVarP(&id, "holder-identity", "", os.Getenv("POD_NAME"), "the holder identity name")
	flag.StringVarP(&driverName, "vm-driver", "d", "docker", "driver is one of: [fake docker]")
	flag.BoolVarP(&leaderElect, "leader-elect", "", false, "start a leader election client and gain leadership before executing the main loop")
	flag.StringVar(&leaseLockName, "lease-lock-name", "vm-crd-create", "the lease lock resource name")
	flag.StringVar(&leaseLockNamespace, "lease-lock-namespace", "", "the lease lock resource namespace")
	flag.BoolVarP(&showVersion, "version", "", false, "display the version")
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
}

func restConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {

	// klog.InitFlags(nil)
	parseFlags()

	if showVersion {
		fmt.Fprintf(os.Stdout, "%s\n", version.GetVersion())
		os.Exit(0)
	}

	k8scfg, err := restConfig(kubeconfig)
	if err != nil {
		klog.Fatalf("Error to build rest config: %s", err.Error())
	}

	k8sclientset := kubernetes.NewForConfigOrDie(k8scfg)

	clientset, err := cloudnative.NewForConfig(k8scfg)
	if err != nil {
		klog.Fatalf("Error to build cloudnative clientset: %s", err.Error())
	}

	var vmDriver driver.Interface
	switch driverName {
	case "docker":
		docker, err := driver.NewDockerDriver()
		if err != nil {
			klog.Fatalf("Error to new docker driver: %s", err.Error())
		}
		vmDriver = docker
	default:
		klog.Fatalf("The driver '%s' is not supported.", driverName)
	}

	informer := cloudnativeinformer.NewSharedInformerFactory(clientset, defaultSyncTime)
	controller := controller.New(clientset, informer, vmDriver)

	// use a Go context so we can tell the leaderelection code when we
	// want to step down
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// listen for interrupts or the Linux SIGTERM signal and cancel
	// our context, which the leader election code will observe and
	// step down
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		cancel()
	}()

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	informer.Start(ctx.Done())

	if leaderElect {
		lock := &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Name:      leaseLockName,
				Namespace: leaseLockNamespace,
			},
			Client: k8sclientset.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: id,
			},
		}
		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock:            lock,
			ReleaseOnCancel: true,
			LeaseDuration:   30 * time.Second,
			RenewDeadline:   15 * time.Second,
			RetryPeriod:     5 * time.Second,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					if err := controller.Run(ctx.Done(), threads); err != nil {
						klog.Fatalf("Error to run the controller instance: %s.", err)
					}
					klog.Infof("%s: leading", id)
				},
				OnStoppedLeading: func() {
					controller.Stop()
					klog.Infof("%s: lost", id)
				},
				OnNewLeader: func(identity string) {
					if identity == id {
						return
					}
					klog.Infof("new leader elected: %s", identity)
				},
			},
		})
	} else {
		if err := controller.Run(ctx.Done(), threads); err != nil {
			klog.Fatalf("Error to run the controller instance: %s.", err)
		}
	}
}

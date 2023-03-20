/*
Copyright © 2023 The vm-crd-create Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

type CreateRequest struct {
	Name   string
	CPU    int64
	Memory int64
}

type CreateReponse struct {
	ID string
}

type GetStatusResponse struct {
	State            string
	CPUPercentage    float64
	MemoryPercentage float64
}

type Interface interface {
	CreateServer(*CreateRequest) (*CreateReponse, error)
	DeleteServer(name string) error
	IsServerExist(name string) (bool, error)
	GetServerStatus(name string) (*GetStatusResponse, error)
}

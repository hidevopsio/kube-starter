// Copyright 2018 John Deng (hi.devops.io@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package helloworld provides the quick start web application example
// main package
package main

// import web starter from hiboot
import (
	goctx "context"
	"github.com/hidevopsio/hiboot/pkg/app/web/context"
	"github.com/hidevopsio/hiboot/pkg/log"
	"github.com/hidevopsio/kube-starter/pkg/oidc"

	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/app/web"
	"github.com/hidevopsio/hiboot/pkg/at"
	"github.com/hidevopsio/hiboot/pkg/model"
	"github.com/hidevopsio/hiboot/pkg/starter/actuator"
	"github.com/hidevopsio/hiboot/pkg/starter/swagger"
	"github.com/hidevopsio/kube-starter/pkg/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Controller Rest Controller with path /
// RESTful Controller, derived from at.RestController. The context mapping of this controller is '/' by default
type Controller struct {
	// at.RestController or at.RestController must be embedded here
	at.RestController
	at.RequestMapping `value:"/api/v1/namespaces/{namespace}"`
}

func newController() *Controller {
	log.Infof("newController")
	return &Controller{}
}

type PodListResponse struct {
	model.BaseResponseInfo
	Data *corev1.PodList `json:"data"`
}

// ListPods list all pods for specific namespace
func (c *Controller) ListPods(_ struct {
	at.GetMapping `value:"/pods"`
	at.Operation  `id:"List Pods" description:"List Pods of giving namespace"`
	at.Consumes   `values:"application/json"`
	at.Produces   `values:"application/json"`
	Parameters    struct {
		at.Parameter `type:"string" name:"namespace" in:"path" description:"Path Variable（Namespace）" required:"true"`
	}
	Responses struct {
		StatusOK struct {
			at.Parameter `name:"Namespace" in:"body" description:"Get Pod List"`
			PodListResponse
		}
	}
}, namespace string, fnClient kubeclient.FnClient) (response *PodListResponse, err error) {
	response = new(PodListResponse)
	var podList corev1.PodList

	err = fnClient().List(goctx.TODO(), &podList, client.InNamespace(namespace))
	if err == nil {
		response.Data = &podList
	}

	// response
	return
}

// ListPodsByUser list all pods by user
func (c *Controller) ListPodsByUser(_ struct {
	at.GetMapping `value:"/pods/user"`
	at.Operation  `id:"List Pods" description:"List Pods of giving namespace"`
	at.Consumes   `values:"application/json"`
	at.Produces   `values:"application/json"`
	Parameters    struct {
		at.Parameter `type:"string" name:"namespace" in:"path" description:"Path Variable（Namespace）" required:"true"`
	}
	Responses struct {
		StatusOK struct {
			at.Parameter `name:"Namespace" in:"body" description:"Get Pod List"`
			PodListResponse
		}
	}
}, namespace string, fnRuntimeClient kubeclient.FnRuntimeClient, ctx context.Context) (response *PodListResponse, err error) {
	response = new(PodListResponse)
	var podList corev1.PodList

	err = fnRuntimeClient(ctx).List(goctx.TODO(), &podList, client.InNamespace(namespace))
	if err == nil {
		response.Data = &podList
	}

	return
}

type ServiceListResponse struct {
	model.BaseResponseInfo
	Data *corev1.ServiceList `json:"data"`
}

// ListServices list all services
func (c *Controller) ListServices(_ struct {
	at.GetMapping `value:"/services"`
	at.Operation  `id:"List Services" description:"List Services of giving namespace"`
	at.Consumes   `values:"application/json"`
	at.Produces   `values:"application/json"`
	Parameters    struct {
		at.Parameter `type:"string" name:"namespace" in:"path" description:"Path Variable（Namespace）" required:"true"`
	}
	Responses struct {
		StatusOK struct {
			at.Parameter `name:"Namespace" in:"body" description:"Get Service List"`
			ServiceListResponse
		}
	}
}, namespace string, fnRuntimeClient kubeclient.FnRuntimeClient, ctx context.Context) (response *ServiceListResponse, err error) {
	response = new(ServiceListResponse)
	var serviceList corev1.ServiceList

	err = fnRuntimeClient(ctx).List(goctx.TODO(), &serviceList, client.InNamespace(namespace))
	if err == nil {
		response.Data = &serviceList
	}

	return
}

type DeploymentListResponse struct {
	model.BaseResponseInfo
	Data *appsv1.DeploymentList `json:"data"`
}

// ListDeployment list all deployments
func (c *Controller) ListDeployment(_ struct {
	at.GetMapping `value:"/deployments"`
	at.Operation  `id:"List Deployments" description:"List Deployments of giving namespace"`
	at.Consumes   `values:"application/json"`
	at.Produces   `values:"application/json"`
	Parameters    struct {
		at.Parameter `type:"string" name:"namespace" in:"path" description:"Path Variable（Namespace）" required:"true"`
	}
	Responses struct {
		StatusOK struct {
			at.Parameter `name:"Namespace" in:"body" description:"Get Deployment List"`
			PodListResponse
		}
	}
}, namespace string, token *oidc.Token, fnClient kubeclient.FnClient) (response *DeploymentListResponse, err error) {
	response = new(DeploymentListResponse)
	var deploymentList appsv1.DeploymentList

	err = fnClient().List(goctx.TODO(), &deploymentList, client.InNamespace(namespace))
	if err == nil {
		user := "unknown"
		if token.Claims != nil {
			user = token.Claims.Subject
		}
		response.Message = user + " Got Deployment List"
		response.Data = &deploymentList
	}

	// response
	return
}

// main function
func main() {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	app.Register(
		scheme,
		newController,
		swagger.ApiInfoBuilder().
			Title("HiBoot Example - Hello world").
			Description("This is an example that demonstrate the basic usage"))

	// create new web application and run it
	web.NewApplication().
		SetProperty(app.ProfilesInclude, swagger.Profile, web.Profile, actuator.Profile, kubeclient.Profile, oidc.Profile).
		SetProperty("logging.level", "debug").
		Run()
}

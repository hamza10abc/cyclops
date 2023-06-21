package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cyclops-ui/cycops-ctrl/internal/cluster/k8sclient"
	"github.com/cyclops-ui/cycops-ctrl/internal/mapper"
	"github.com/cyclops-ui/cycops-ctrl/internal/models/dto"
	"github.com/cyclops-ui/cycops-ctrl/internal/storage/templates"
	"github.com/cyclops-ui/cycops-ctrl/internal/template"
)

type Modules struct {
	kubernetesClient *k8sclient.KubernetesClient
	templates        *templates.Storage
}

func NewModulesController(templates *templates.Storage, kubernetes *k8sclient.KubernetesClient) *Modules {
	return &Modules{
		kubernetesClient: kubernetes,
		templates:        templates,
	}
}

func (m *Modules) GetModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, mapper.ModuleToDTO(*module))
}

func (m *Modules) ListModules(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	modules, err := m.kubernetesClient.ListModules()
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, mapper.ModuleListToDTO(modules))
}

func (m *Modules) DeleteModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	err := m.kubernetesClient.DeleteModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func (m *Modules) DeleteModuleResource(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	var request dto.DeleteResource
	if err := ctx.BindJSON(&request); err != nil {
		fmt.Println("error binding request", request)
		ctx.Status(http.StatusBadRequest)
		return
	}

	err := m.kubernetesClient.Delete(&request)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func (m *Modules) CreateModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	var request dto.Module
	if err := ctx.BindJSON(&request); err != nil {
		fmt.Println("error binding request", request)
		ctx.JSON(http.StatusBadRequest, dto.NewError("Error loading template", err.Error()))
		return
	}

	err := m.kubernetesClient.CreateModule(mapper.RequestToModule(request))
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error creating module", err.Error()))
		return
	}

	ctx.Status(http.StatusOK)
}

func (m *Modules) UpdateModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	var request dto.Module
	if err := ctx.BindJSON(&request); err != nil {
		fmt.Println("error binding request", request)
		ctx.Status(http.StatusBadRequest)
		return
	}

	curr, err := m.kubernetesClient.GetModule(request.Name)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	module := mapper.RequestToModule(request)
	module.SetResourceVersion(curr.GetResourceVersion())

	err = m.kubernetesClient.UpdateModule(module)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func (m *Modules) ResourcesForModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	template, err := m.templates.GetConfig(module.Spec.TemplateRef)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	resources, err := m.kubernetesClient.GetResourcesForModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	resources, err = m.kubernetesClient.GetDeletedResources(resources, *module, template)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, resources)
}

func (m *Modules) Template(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	currentTemplate, err := m.templates.GetConfig(module.Spec.TemplateRef)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	currentManifest, err := template.HelmTemplate(*module, currentTemplate)
	if err != nil {
		fmt.Println("error templating current", err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	proposedTemplate, err := m.templates.GetConfig(module.Spec.TemplateRef)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	proposedManifest, err := template.HelmTemplate(*module, proposedTemplate)
	if err != nil {
		fmt.Println("error templating current", err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	res := dto.TemplatesResponse{
		Current: currentManifest,
		New:     proposedManifest,
	}

	ctx.JSON(http.StatusOK, res)
}

func (m *Modules) HelmTemplate(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	currentTemplate, err := m.templates.GetConfig(module.Spec.TemplateRef)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	_, err = template.HelmTemplate(*module, currentTemplate)
	if err != nil {
		fmt.Println("error templating current", err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, "{}")
}

//func (m *Modules) ModuleToResources(ctx *gin.Context) {
//	err := m.kubernetesClient.ModuleToResources("test")
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	ctx.Header("Access-Control-Allow-Origin", "*")
//	ctx.Status(http.StatusOK)
//}
//
//func (m *Modules) ResourcesForModule(ctx *gin.Context) {
//	resources, err := m.kubernetesClient.ResourcesForModule(ctx.Param("name"))
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	ctx.Header("Access-Control-Allow-Origin", "*")
//	ctx.JSON(http.StatusOK, resources)
//}

func (m *Modules) GetLogs(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	logs, err := m.kubernetesClient.GetPodLogs(ctx.Param("namespace"), ctx.Param("container"), ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, logs)
}
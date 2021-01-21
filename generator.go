package main

import (
	"fmt"
	"github.com/go-kratos/kratos/tool/protobuf/pkg/generator"
	"github.com/go-kratos/kratos/tool/protobuf/pkg/naming"
	"github.com/go-kratos/kratos/tool/protobuf/pkg/tag"
	"github.com/go-kratos/kratos/tool/protobuf/pkg/typemap"
	"github.com/go-kratos/kratos/tool/protobuf/pkg/utils"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"reflect"
	"sort"
	"strings"
)

type gin struct {
	generator.Base
	filesHandled int
}

func GinGenerator() *gin {
	t := &gin{}
	return t
}

// Generate ...
func (t *gin) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	t.Setup(in)

	// Showtime! Generate the response.
	resp := new(plugin.CodeGeneratorResponse)
	for _, f := range t.GenFiles {
		respFile := t.generateForFile(f)
		if respFile != nil {
			resp.File = append(resp.File, respFile)
		}
	}
	return resp
}

func (t *gin) generateForFile(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	resp := new(plugin.CodeGeneratorResponse_File)

	t.generateFileHeader(file, t.GenPkgName)
	t.generateImports(file)
	t.generatePathConstants(file)
	count := 0
	for _, service := range file.Service {
		count += t.generateGinInterface(file, service)
		t.generateGinRoute(file, service)
	}

	resp.Name = proto.String(naming.GenFileName(file, ".gin.go"))
	resp.Content = proto.String(t.FormattedOutput())
	t.Output.Reset()

	t.filesHandled++
	return resp
}

func (t *gin) generatePathConstants(file *descriptor.FileDescriptorProto) {
	t.P()
	for _, service := range file.Service {
		name := naming.ServiceName(service)
		for _, method := range service.Method {
			if !t.ShouldGenForMethod(file, service, method) {
				continue
			}
			apiInfo := t.GetHttpInfoCached(file, service, method)
			t.P(`var Path`, name, naming.MethodName(method), ` = "`, apiInfo.Path, `"`)
		}
		t.P()
	}
}

func (t *gin) generateFileHeader(file *descriptor.FileDescriptorProto, pkgName string) {
	t.P("// Code generated by protoc-gen-gin.", " DO NOT EDIT.")
	t.P("// source: ", file.GetName())
	t.P()
	t.P(`package `, pkgName)
	t.P()
}

func (t *gin) generateImports(file *descriptor.FileDescriptorProto) {
	//if len(file.Service) == 0 {
	//	return
	//}
	t.P(`import (`)
	//t.P(`	`,t.pkgs["context"], ` "context"`)
	t.P(`	"context"`)
	t.P(`	"net/http"`)
	t.P()
	t.P(`	"github.com/gin-gonic/gin"`)
	t.P(`	"github.com/gin-gonic/gin/binding"`)
	t.P(`	"github.com/liu-junyong/protoc-gen-gin/ecode"`)
	t.P(`)`)
	// It's legal to import a message and use it as an input or output for a
	// method. Make sure to import the package of any such message. First, dedupe
	// them.
	deps := t.DeduceDeps(file)
	for pkg, importPath := range deps {
		t.P(`import `, pkg, ` `, importPath)
	}
	t.P()
	t.P(`// to suppressed 'imported but not used warning'`)
	t.P(`var _ *gin.Context`)
	t.P(`var _ context.Context`)
	t.P(`var _ binding.StructValidator`)
}

// Big header comments to makes it easier to visually parse a generated file.
func (t *gin) sectionComment(sectionTitle string) {
	t.P()
	t.P(`// `, strings.Repeat("=", len(sectionTitle)))
	t.P(`// `, sectionTitle)
	t.P(`// `, strings.Repeat("=", len(sectionTitle)))
	t.P()
}

func (t *gin) generateGinRoute(
	file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto) {
	t.P()
	t.P(`func JSON(c *gin.Context, data interface{}, err error) {
		httpCode := http.StatusOK
		bcode := ecode.Cause(err)
		if bcode.Code < 0 {
			httpCode = -bcode.Code
		}
		c.JSON(httpCode, Response{
			Code:    bcode.Code,
			Message: bcode.Message,
			Data:    data,
		})
	}`)
	t.P()
	t.P(`type Response struct {`)
	t.P("Code int" + " `" + `json:"code"` + "`")
	t.P("Message string" + " `" + `json:"message"` + "`")
	t.P("Data interface{}" + " `" + `json:"data,omitempty"` + "`")
	t.P(`}`)
	t.P()

	// old mode is generate xx.route.go in the http pkg
	// new mode is generate route code in the same .bm.go
	// route rule /x{department}/{project-name}/{path_prefix}/method_name
	// generate each route method
	servName := naming.ServiceName(service)
	versionPrefix := naming.GetVersionPrefix(t.GenPkgName)
	svcName := utils.LcFirst(utils.CamelCase(versionPrefix)) + servName + "Svc"
	t.P(`var `, svcName, ` `, servName, `GinServer`)

	type methodInfo struct {
		middlewares   []string
		routeFuncName string
		apiInfo       *generator.HTTPInfo
		methodName    string
	}
	var methList []methodInfo
	var allMiddlewareMap = make(map[string]bool)
	for _, method := range service.Method {
		if !t.ShouldGenForMethod(file, service, method) {
			continue
		}
		var middlewares []string
		comments, _ := t.Reg.MethodComments(file, service, method)
		tags := tag.GetTagsInComment(comments.Leading)
		if tag.GetTagValue("dynamic", tags) == "true" {
			continue
		}
		apiInfo := t.GetHttpInfoCached(file, service, method)
		midStr := tag.GetTagValue("middleware", tags)
		if midStr != "" {
			middlewares = strings.Split(midStr, ",")
			for _, m := range middlewares {
				allMiddlewareMap[m] = true
			}
		}

		methName := naming.MethodName(method)
		inputType := t.GoTypeName(method.GetInputType())

		routeName := utils.LcFirst(utils.CamelCase(servName) +
			utils.CamelCase(methName))

		methList = append(methList, methodInfo{
			apiInfo:       apiInfo,
			middlewares:   middlewares,
			routeFuncName: routeName,
			methodName:    method.GetName(),
		})

		t.P(fmt.Sprintf("func %s (c *gin.Context) {", routeName))
		t.P(`	p := new(`, inputType, `)`)
		if t.hasHeaderTag(t.Reg.MessageDefinition(method.GetInputType())) {
		}
		t.P(``)
		t.P(`	if err := c.ShouldBind(p) ; err != nil {`)
		t.P(`		return`)
		t.P(`	}`)
		t.P(`	resp, err := `, svcName, `.`, methName, `(c, p)`)
		t.P(`	JSON(c, resp, err)`)
		t.P(`}`)
		t.P()
	}

	// generate route group
	var midList []string
	for m := range allMiddlewareMap {
		midList = append(midList, m+" gin.HandlerFunc")
	}

	sort.Strings(midList)

	// 新的注册路由的方法
	var ginFuncName = fmt.Sprintf("Register%sGinServer", servName)
	t.P(`// `, ginFuncName, ` Register the gin route`)
	t.P(`func `, ginFuncName, `(c *gin.Engine, server `, servName, `GinServer) {`)
	t.P(svcName, ` = server`)
	for _, methInfo := range methList {
		t.P(`c.`, methInfo.apiInfo.HttpMethod, `("`, methInfo.apiInfo.Path, `",`, methInfo.routeFuncName, ` )`)
	}
	t.P(`	}`)
}

func (t *gin) hasHeaderTag(md *typemap.MessageDefinition) bool {
	if md.Descriptor.Field == nil {
		return false
	}
	for _, f := range md.Descriptor.Field {
		t := tag.GetMoreTags(f)
		if t != nil {
			st := reflect.StructTag(*t)
			if st.Get("request") != "" {
				return true
			}
			if st.Get("header") != "" {
				return true
			}
		}
	}
	return false
}

func (t *gin) generateGinInterface(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) int {
	count := 0
	servName := naming.ServiceName(service)
	t.P("// " + servName + "GinServer is the server API for " + servName + " service.")

	comments, err := t.Reg.ServiceComments(file, service)
	if err == nil {
		t.PrintComments(comments)
	}
	t.P(`type `, servName, `GinServer interface {`)
	for _, method := range service.Method {
		if !t.ShouldGenForMethod(file, service, method) {
			continue
		}
		count++
		t.generateInterfaceMethod(file, service, method, comments)
		t.P()
	}
	t.P(`}`)
	return count
}

func (t *gin) generateInterfaceMethod(file *descriptor.FileDescriptorProto,
	service *descriptor.ServiceDescriptorProto,
	method *descriptor.MethodDescriptorProto,
	comments typemap.DefinitionComments) {
	comments, err := t.Reg.MethodComments(file, service, method)

	methName := naming.MethodName(method)
	outputType := t.GoTypeName(method.GetOutputType())
	inputType := t.GoTypeName(method.GetInputType())
	tags := tag.GetTagsInComment(comments.Leading)
	if tag.GetTagValue("dynamic", tags) == "true" {
		return
	}

	if err == nil {
		t.PrintComments(comments)
	}

	respDynamic := tag.GetTagValue("dynamic_resp", tags) == "true"
	if respDynamic {
		t.P(fmt.Sprintf(`	%s(c *gin.Context, req *%s) (resp interface{}, err error)`,
			methName, inputType))
	} else {
		t.P(fmt.Sprintf(`	%s(c *gin.Context, req *%s) (resp *%s, err error)`,
			methName, inputType, outputType))
	}
}

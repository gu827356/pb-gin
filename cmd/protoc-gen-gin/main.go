package main

import (
	"flag"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gu827356/pb-gin/pb_gen/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

var requireUnimplemented *bool
var needSwag *bool

const (
	ginPackage         = protogen.GoImportPath("github.com/gin-gonic/gin")
	ginsPackage        = protogen.GoImportPath("github.com/gu827356/pb-gin/gins")
	annotationsPackage = protogen.GoImportPath("github.com/gu827356/pb-gin/pb_gen/googleapis/api/annotations")
	jsonpbPackage      = protogen.GoImportPath("github.com/golang/protobuf/jsonpb")
)

var jsonpbMarshaler = jsonpb.Marshaler{}

func main() {
	var flags flag.FlagSet
	requireUnimplemented = flags.Bool("require_unimplemented_servers", true, "set to false to match legacy behavior")
	needSwag = flags.Bool("swag", false, "set to true to generate swag annotations")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			generateFile(gen, f)
		}
		return nil
	})
}

func generateFile(gen *protogen.Plugin, f *protogen.File) {
	if len(f.Services) == 0 {
		return
	}
	filename := f.GeneratedFilenamePrefix + "gin_service.pb.go"

	g := gen.NewGeneratedFile(filename, f.GoImportPath)
	g.P("// Code generated by protoc-gen-protoc-gen-gin. DO NOT EDIT.")
	g.P()
	g.P("package ", f.GoPackageName)
	g.P("")

	generateFileContent(gen, f, g)
}

func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	for _, service := range file.Services {
		generateService(gen, file, g, service)
		g.P()
	}
}

func generateService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	g.P("type ", goServerType(service), " interface {")
	for _, method := range service.Methods {
		g.P(generateClientMethodSignature(g, method))
	}
	g.P("}")
	g.P()

	generateRegisterFunc(g, service)
}

func generateClientMethodSignature(g *protogen.GeneratedFile, method *protogen.Method) string {
	s := method.GoName + "(c *" + g.QualifiedGoIdent(ginPackage.Ident("Context"))
	s += ", in *" + g.QualifiedGoIdent(method.Input.GoIdent)
	s += ") ("
	s += "*" + g.QualifiedGoIdent(method.Output.GoIdent)
	s += ", error)"
	return s
}

func generateRegisterFunc(g *protogen.GeneratedFile, service *protogen.Service) {
	g.P("func Register", goServerType(service), "(r ", g.QualifiedGoIdent(ginPackage.Ident("IRouter")), ", svr ", goServerType(service), ") {")
	g.P("register := ", goRegisterType(service), "{svr:svr, r:r}")
	g.P("register.register()")
	g.P("}")
	g.P()

	generateRegister(g, service)
}

func generateRegister(g *protogen.GeneratedFile, service *protogen.Service) {
	g.P("type ", goRegisterType(service), " struct {")
	g.P("svr ", goServerType(service))
	g.P("r ", g.QualifiedGoIdent(ginPackage.Ident("IRouter")))
	g.P("}")
	g.P()

	g.P("func (rr *", goRegisterType(service), ") register() {")
	for _, method := range service.Methods {
		g.P("rr.register", method.GoName, "()")
	}
	g.P("}")
	g.P()

	for _, method := range service.Methods {
		generateRegisterMethod(g, service, method)
		g.P()
	}

	g.P("func (rr *", goRegisterType(service), ") unmarshalHttpRule(s string) *", g.QualifiedGoIdent(annotationsPackage.Ident("HttpRule")), " {")
	g.P("ret := ", g.QualifiedGoIdent(annotationsPackage.Ident("HttpRule")), "{}")
	g.P("_ = ", jsonpbPackage.Ident("UnmarshalString"), "(s, &ret)")
	g.P("return &ret")
	g.P("}")
	g.P()
}

func generateRegisterMethod(g *protogen.GeneratedFile, service *protogen.Service, method *protogen.Method) {
	rule := getHttpRule(method)
	ruleJson, _ := jsonpbMarshaler.MarshalToString(rule)

	if *needSwag {
		generateSwag(g, service, method, rule)
	}

	g.P("func (rr *", goRegisterType(service), ") register", method.GoName, "() {")

	g.P("rule := rr.unmarshalHttpRule(`", ruleJson, "`)")
	g.P(g.QualifiedGoIdent(ginsPackage.Ident("RegisterRoute")), "(rr.r, rule, func(c *", ginPackage.Ident("Context"), ")(interface{}, error) {")
	g.P("req := ", g.QualifiedGoIdent(method.Input.GoIdent), "{}")
	g.P("if err := ", g.QualifiedGoIdent(ginsPackage.Ident("ParseRequest")), "(c, &req, rule); err != nil {")
	g.P("return nil, err")
	g.P("}")
	g.P("return rr.svr.", method.GoName, "(c, &req)")
	g.P("})")

	g.P("}")
}

func generateSwag(g *protogen.GeneratedFile, service *protogen.Service, method *protogen.Method, rule *annotations.HttpRule) {
	var httpMethod string
	var path string
	switch p := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		g.P("// @Accept x-www-form-urlencoded")
		httpMethod = "get"
		path = p.Get
	case *annotations.HttpRule_Post:
		g.P("// @Accept json")
		httpMethod = "post"
		path = p.Post
	case *annotations.HttpRule_Put:
		g.P("// @Accept json")
		httpMethod = "put"
		path = p.Put
	default:
		return
	}
	g.P("// @Produce json")

	if httpMethod == "get" {
		for _, field := range method.Input.Fields {
			g.P("// @Param ", field.Desc.Name(), ` query string false ""`)
		}
	}

	g.P("// @Success 200 {object} ", g.QualifiedGoIdent(method.Output.GoIdent), ` ""`)
	g.P("// @Router ", path, " [", httpMethod, "]")
}

func getHttpRule(method *protogen.Method) *annotations.HttpRule {
	options := method.Desc.Options()
	var httpRule *annotations.HttpRule
	if proto.HasExtension(options, annotations.E_Http) {
		httpRule = proto.GetExtension(options, annotations.E_Http).(*annotations.HttpRule)
	}
	if httpRule == nil {
		httpRule = &annotations.HttpRule{}
	}
	return httpRule
}

func goServerType(service *protogen.Service) string {
	return service.GoName + "Server"
}

func goRegisterType(service *protogen.Service) string {
	return "register" + goServerType(service)
}

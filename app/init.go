package app

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"io/ioutil"
	"path/filepath"
	"github.com/revel/revel"
	"k8s-devops-console/app/models"
	"github.com/revel/revel/logger"
	"k8s.io/apimachinery/pkg/runtime"
		"k8s.io/api/core/v1"
	"k8s.io/api/settings/v1alpha1"
	v13 "k8s.io/api/rbac/v1"
	v12 "k8s.io/api/networking/v1"
	)

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string
)

const (
	DEFAULT_NAMESPACE_FILTER_ACCESS = `^.*$`
	DEFAULT_NAMESPACE_FILTER_DELETE = `^.*$`
	DEFAULT_NAMESPACE_FILTER_USER = `^user-%s-`
	DEFAULT_NAMESPACE_FILTER_TEAM = `^team-%s-`
	NAMESPACE_ENVIRONMENTS = "dev,test,int,load,prod,team,user"
	NAMESPACE_TEAM = `^[a-zA-Z0-9]{3,}$`
	NAMESPACE_APP  = `^[a-zA-Z0-9]{3,}$`
)

var (
	RegexpNamespaceEnv *regexp.Regexp
	RegexpNamespaceTeam *regexp.Regexp
	RegexpNamespaceApp *regexp.Regexp
	RegexpNamespaceFilter *regexp.Regexp
	RegexpNamespaceDeleteFilter *regexp.Regexp
	NamespaceEnvironments []string
	NamespaceFilterUser string
	NamespaceFilterTeam string
	AppConfig *models.AppConfig
	AuditLog logger.MultiLogger
	KubeNamespaceConfig map[string]*KubeObjectList
)

type KubeObjectList struct {
	ConfigMaps map[string]KubeObject
	ServiceAccounts map[string]KubeObject
	Roles map[string]KubeObject
	RoleBindings map[string]KubeObject
	PodPresets map[string]KubeObject
	NetworkPolicies map[string]KubeObject
	LimitRanges map[string]KubeObject
	ResourceQuotas map[string]KubeObject
}

type KubeObject struct {
	Name string
	Path string
	Object runtime.Object
}


func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.ActionInvoker,           // Invoke the action.
	}

	logger.LogFunctionMap["stdoutjson"]=
		func(c *logger.CompositeMultiHandler, options *logger.LogOptions) {
			// Set the json formatter to os.Stdout, replace any existing handlers for the level specified
			c.SetJson(os.Stdout, options)
		}

	logger.LogFunctionMap["stderrjson"]=
		func(c *logger.CompositeMultiHandler, options *logger.LogOptions) {
			// Set the json formatter to os.Stdout, replace any existing handlers for the level specified
			c.SetJson(os.Stderr, options)
		}


	// Register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	// revel.OnAppStart(ExampleStartupScript)
	// revel.OnAppStart(InitDB)
	// revel.OnAppStart(FillCache)
	revel.OnAppStart(InitLogger)
	revel.OnAppStart(InitConfig)
	revel.OnAppStart(InitTemplateEngine)
	revel.OnAppStart(InitAppConfiguration)
	revel.OnAppStart(InitKubeNamespaceConfig)
}

// HeaderFilter adds common security headers
// There is a full implementation of a CSRF filter in
// https://github.com/revel/modules/tree/master/csrf
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

func GetConfigString(key, defaultValue string) (ret string) {
	ret = defaultValue

	// try to get config
	if val, exists := revel.Config.String(key); exists && val != "" {
		return val
	}

	// try to get config default
	if val, exists := revel.Config.String(key + ".default"); exists && val != "" {
		return val
	}

	return
}

func GetConfigInt(key string, defaultValue int) (ret int) {
	ret = defaultValue

	// try to get config
	if val, exists := revel.Config.String(key); exists && val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}

	// try to get config default
	if val, exists := revel.Config.String(key + ".default"); exists && val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}

	return
}

func InitLogger() {
	AuditLog = revel.AppLog.New("system", "audit")
}

func InitConfig() {
	RegexpNamespaceFilter = regexp.MustCompile(GetConfigString("k8s.namespace.access.filter", DEFAULT_NAMESPACE_FILTER_ACCESS))
	RegexpNamespaceDeleteFilter = regexp.MustCompile(GetConfigString("k8s.namespace.delete.filter", DEFAULT_NAMESPACE_FILTER_DELETE))
	RegexpNamespaceTeam = regexp.MustCompile(GetConfigString("k8s.namespace.validation.team", NAMESPACE_TEAM))
	RegexpNamespaceApp = regexp.MustCompile(GetConfigString("k8s.namespace.validation.app", NAMESPACE_APP))
	NamespaceFilterUser = GetConfigString("k8s.namespace.user.filter.", DEFAULT_NAMESPACE_FILTER_USER)
	NamespaceFilterTeam = GetConfigString("k8s.namespace.team.filter", DEFAULT_NAMESPACE_FILTER_TEAM)

	envList := GetConfigString("k8s.namespace.environments", NAMESPACE_ENVIRONMENTS)
	NamespaceEnvironments = strings.Split(envList, ",")
}

func InitTemplateEngine() {
	revel.TemplateFuncs["config"] = func(option string) string {
		return GetConfigString(option, "")
	}
}

func InitAppConfiguration() {
	var appYamlPath string
	for _, path := range revel.ConfPaths {
		path = filepath.Join(path, GetConfigString("k8s.config", "app.yaml"))
		if _, err := os.Stat(path); err == nil {
			appYamlPath = path
		}
	}

	if appYamlPath != "" {
		data, err := ioutil.ReadFile(appYamlPath)
		if err != nil {
			panic(err)
		}

		AppConfig, err = models.AppConfigCreateFromYaml(string(data))
		if err != nil {
			panic(err)
		}
	} else {
		AppConfig = &models.AppConfig{}
	}
}

func createKubeObjectList() (list *KubeObjectList) {
	list = &KubeObjectList{}
	list.ConfigMaps = map[string]KubeObject{}
	list.ServiceAccounts = map[string]KubeObject{}
	list.Roles = map[string]KubeObject{}
	list.RoleBindings = map[string]KubeObject{}
	list.ResourceQuotas = map[string]KubeObject{}
	list.NetworkPolicies = map[string]KubeObject{}
	list.PodPresets = map[string]KubeObject{}
	list.LimitRanges = map[string]KubeObject{}
	return
}

func InitKubeNamespaceConfig() {
	var k8sYamlPath string
	for _, path := range revel.ConfPaths {
		path = filepath.Join(path, "k8s")
		if _, err := os.Stat(path); err == nil {
			k8sYamlPath = path
		}
	}

	KubeNamespaceConfig = map[string]*KubeObjectList{}

	if k8sYamlPath != "" {
		// default namespace settings
		k8sDefaultPath := filepath.Join(k8sYamlPath, "_default")
		if (!IsDirectory(k8sDefaultPath)) {
			k8sDefaultPath = ""
		}
		KubeNamespaceConfig["_default"] = buildKubeConfigList("", k8sDefaultPath)

		// parse config for each subpath as environment
		err := filepath.Walk(k8sYamlPath, func(path string, info os.FileInfo, err error) error {
			// jump into base dir
			if path == k8sYamlPath {
				return nil
			}

			// parse configs in dir but don't jump recursive into it
			if info.IsDir() && path != k8sDefaultPath {
				KubeNamespaceConfig[info.Name()] = buildKubeConfigList(k8sDefaultPath, path)
				return filepath.SkipDir
			}
			return nil
		})

		if err != nil {
			panic(err)
		}
	}
}

func buildKubeConfigList(defaultPath, path string) (*KubeObjectList) {
	kubeConfigList := createKubeObjectList()

	if defaultPath != "" {
		addK8sConfigsFromPath(defaultPath, kubeConfigList)
	}

	if path != "" {
		addK8sConfigsFromPath(path, kubeConfigList)
	}

	return kubeConfigList
}

func addK8sConfigsFromPath(configPath string, list *KubeObjectList) {
	var fileList []string
	filepath.Walk(configPath, func(path string, f os.FileInfo, err error) error {
		if IsK8sConfigFile(path) {
			fileList = append(fileList, path)
		}
		return nil
	})


	for _, path := range fileList {
		item := KubeObject{}
		item.Path = path
		item.Object = KubeParseConfig(path)

		switch(item.Object.GetObjectKind().GroupVersionKind().Kind) {
		case "ConfigMap":
			item.Name = item.Object.(*v1.ConfigMap).Name
			list.ConfigMaps[item.Name] = item
		case "ServiceAccount":
			item.Name = item.Object.(*v1.ServiceAccount).Name
			list.ServiceAccounts[item.Name] = item
		case "Role":
			item.Name = item.Object.(*v13.Role).Name
			list.Roles[item.Name] = item
		case "RoleBinding":
			item.Name = item.Object.(*v13.RoleBinding).Name
			list.RoleBindings[item.Name] = item
		case "NetworkPolicy":
			item.Name = item.Object.(*v12.NetworkPolicy).Name
			list.NetworkPolicies[item.Name] = item
		case "LimitRange":
			item.Name = item.Object.(*v1.LimitRange).Name
			list.LimitRanges[item.Name] = item
		case "PodPreset":
			item.Name = item.Object.(*v1alpha1.PodPreset).Name
			list.PodPresets[item.Name] = item
		case "ResourceQuota":
			item.Name = item.Object.(*v1.ResourceQuota).Name
			list.ResourceQuotas[item.Name] = item
		default:
			panic("Not allowed object found: " + item.Object.GetObjectKind().GroupVersionKind().Kind)
		}
	}
}

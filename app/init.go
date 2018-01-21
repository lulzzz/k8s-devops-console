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
)

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

	AuditLog.Info("Foobar")
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

//func ExampleStartupScript() {
//	// revel.DevMod and revel.RunMode work here
//	// Use this script to check for dev mode and set dev/prod startup scripts here!
//	if revel.DevMode == true {
//		// Dev mode
//	}
//}

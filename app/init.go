package app

import (
	"github.com/revel/revel"
	"regexp"
	"strings"
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
	NAMESPACE_TEAM   = `^[a-zA-Z0-9]{3,}$`
	NAMESPACE_APP    = `^[a-zA-Z0-9]{3,}$`
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


	// Register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	// revel.OnAppStart(ExampleStartupScript)
	// revel.OnAppStart(InitDB)
	// revel.OnAppStart(FillCache)
	revel.OnAppStart(InitConfig)
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

func InitConfig() {
	RegexpNamespaceFilter = regexp.MustCompile(revel.Config.StringDefault("k8s.namespace.filter.access", DEFAULT_NAMESPACE_FILTER_ACCESS))
	RegexpNamespaceDeleteFilter = regexp.MustCompile(revel.Config.StringDefault("k8s.namespace.filter.delete", DEFAULT_NAMESPACE_FILTER_DELETE))
	RegexpNamespaceTeam = regexp.MustCompile(revel.Config.StringDefault("k8s.namespace.validation.team", NAMESPACE_TEAM))
	RegexpNamespaceApp = regexp.MustCompile(revel.Config.StringDefault("k8s.namespace.validation.app", NAMESPACE_APP))
	NamespaceFilterUser = revel.Config.StringDefault("k8s.namespace.filter.user", DEFAULT_NAMESPACE_FILTER_USER)
	NamespaceFilterTeam = revel.Config.StringDefault("k8s.namespace.filter.team", DEFAULT_NAMESPACE_FILTER_TEAM)

	envList := revel.Config.StringDefault("k8s.namespace.environments", NAMESPACE_ENVIRONMENTS)
	NamespaceEnvironments = strings.Split(envList, ",")
}

//func ExampleStartupScript() {
//	// revel.DevMod and revel.RunMode work here
//	// Use this script to check for dev mode and set dev/prod startup scripts here!
//	if revel.DevMode == true {
//		// Dev mode
//	}
//}

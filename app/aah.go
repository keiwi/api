// GENERATED CODE - DO NOT EDIT
//
// aah framework v0.10 - https://aahframework.org
// FILE: aah.go
// DESC: aah application entry point

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"aahframework.org/aah.v0"
	"aahframework.org/config.v0"
	"aahframework.org/essentials.v0"
	"aahframework.org/log.v0"
	controllers "github.com/keiwi/api/app/controllers"
	models "github.com/keiwi/api/app/models"
	generic_authsec "github.com/keiwi/api/app/security"
)

var (
	// Defining flags
	version    = flag.Bool("version", false, "Display application name, version and build date.")
	configPath = flag.String("config", "", "Absolute path of external config file.")
	profile    = flag.String("profile", "", "Environment profile name to activate. e.g: dev, qa, prod.")
	_          = reflect.Invalid
)

func mergeExternalConfig(e *aah.Event) {
	externalConfig, err := config.LoadFile(*configPath)
	if err != nil {
		log.Fatalf("Unable to load external config: %s", *configPath)
	}

	log.Debug("Merging external config into aah application config")
	if err := aah.AppConfig().Merge(externalConfig); err != nil {
		log.Errorf("Unable to merge external config into aah application[%s]: %s", aah.AppName(), err)
	}
}

func setAppEnvProfile(e *aah.Event) {
	aah.AppConfig().SetString("env.active", *profile)
}



func main() {
	log.Infof("aah framework v%s, requires ≥ go1.8", aah.Version)
	flag.Parse()

	aah.SetAppBuildInfo(&aah.BuildInfo{
		BinaryName: "api.exe",
		Version:    "0.0.1",
		Date:       "2018-04-22T19:46:22+02:00",
	})

	aah.SetAppPackaged(false)

	// display application information
	if *version {
		fmt.Printf("%-12s: %s\n", "Binary Name", aah.AppBuildInfo().BinaryName)
		fmt.Printf("%-12s: %s\n", "Version", aah.AppBuildInfo().Version)
		fmt.Printf("%-12s: %s\n", "Build Date", aah.AppBuildInfo().Date)
		return
	}

	// Apply supplied external config file
	if !ess.IsStrEmpty(*configPath) {
		aah.OnInit(mergeExternalConfig)
	}

	// Apply environment profile
	if !ess.IsStrEmpty(*profile) {
		aah.OnInit(setAppEnvProfile)
	}

	aah.Init("github.com/keiwi/api")

	// Adding all the application controllers which refers 'aah.Context' directly
	// or indirectly from app/controllers/** 
	aah.AddController(
		(*controllers.GroupsController)(nil),
	  []*aah.MethodInfo{
	    &aah.MethodInfo{
	      Name: "RenameGroup",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "rename", Type: reflect.TypeOf((*models.GroupRename)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "CreateGroup",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "create", Type: reflect.TypeOf((*models.GroupCreate)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "EditGroup",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "edit", Type: reflect.TypeOf((*models.EditRequest)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "DeleteGroup",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "delete", Type: reflect.TypeOf((*models.GroupID)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "DeleteGroupWithName",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "delete", Type: reflect.TypeOf((*models.GroupName)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "GetGroups",
	      Parameters: []*aah.ParameterInfo{ 
	      },
	    },&aah.MethodInfo{
	      Name: "ExistsGroup",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "group", Type: reflect.TypeOf((*models.GroupName)(nil))},
	      },
	    },
		},
	)
	aah.AddController(
		(*controllers.UsersController)(nil),
	  []*aah.MethodInfo{
	    &aah.MethodInfo{
	      Name: "UserSignup",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "signup", Type: reflect.TypeOf((*models.User)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "UserLogin",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "login", Type: reflect.TypeOf((*models.User)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "UserInfo",
	      Parameters: []*aah.ParameterInfo{ 
	      },
	    },
		},
	)
	aah.AddController(
		(*controllers.ChecksController)(nil),
	  []*aah.MethodInfo{
	    &aah.MethodInfo{
	      Name: "DeleteCheck",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "delete", Type: reflect.TypeOf((*models.ChecksID)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "GetChecks",
	      Parameters: []*aah.ParameterInfo{ 
	      },
	    },&aah.MethodInfo{
	      Name: "GetCheckWithID",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "check", Type: reflect.TypeOf((*models.ChecksID)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "GetWithClientIDAndCommandID",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "c", Type: reflect.TypeOf((*models.ChecksWithClientCommandID)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "GetWithChecksBetweenDateClient",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "c", Type: reflect.TypeOf((*models.ChecksBetweenDateClient)(nil))},
	      },
	    },
		},
	)
	aah.AddController(
		(*controllers.ClientsController)(nil),
	  []*aah.MethodInfo{
	    &aah.MethodInfo{
	      Name: "CreateClient",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "create", Type: reflect.TypeOf((*models.ClientCreate)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "DeleteClient",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "delete", Type: reflect.TypeOf((*models.ClientID)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "GetClients",
	      Parameters: []*aah.ParameterInfo{ 
	      },
	    },&aah.MethodInfo{
	      Name: "GetClientWithID",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "client", Type: reflect.TypeOf((*models.ClientID)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "EditClient",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "edit", Type: reflect.TypeOf((*models.EditRequest)(nil))},
	      },
	    },
		},
	)
	aah.AddController(
		(*controllers.CommandsController)(nil),
	  []*aah.MethodInfo{
	    &aah.MethodInfo{
	      Name: "CreateCommand",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "create", Type: reflect.TypeOf((*models.CommandCreate)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "EditCommand",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "edit", Type: reflect.TypeOf((*models.EditRequest)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "DeleteCommand",
	      Parameters: []*aah.ParameterInfo{ &aah.ParameterInfo{Name: "delete", Type: reflect.TypeOf((*models.CommandID)(nil))},
	      },
	    },&aah.MethodInfo{
	      Name: "GetCommands",
	      Parameters: []*aah.ParameterInfo{ 
	      },
	    },
		},
	)

	// Initialize application security auth schemes - Authenticator & Authorizer
	secMgr := aah.AppSecurityManager()
	log.Debugf("Calling authenticator Init for auth scheme '%s'", "generic_auth")
	if err := secMgr.GetAuthScheme("generic_auth").SetAuthenticator(&generic_authsec.AuthenticationProvider{}); err != nil {
		log.Fatal(err)
	}
	log.Debugf("Calling authorizer Init for auth scheme '%s'", "generic_auth")
	if err := secMgr.GetAuthScheme("generic_auth").SetAuthorizer(&generic_authsec.AuthorizationProvider{}); err != nil {
		log.Fatal(err)
	}
	

	log.Info("aah application initialized successfully")

	

	go aah.Start()

	// Listen to OS signal's SIGINT & SIGTERM for aah server Shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	sig := <-sc
	switch sig {
	case os.Interrupt:
		log.Warn("Interrupt signal (SIGINT) received")
	case syscall.SIGTERM:
		log.Warn("Termination signal (SIGTERM) received")
	}

	// Call aah shutdown
	aah.Shutdown()
	log.Info("aah application shutdown successful")

	// bye bye, see you later.
	os.Exit(0)
}

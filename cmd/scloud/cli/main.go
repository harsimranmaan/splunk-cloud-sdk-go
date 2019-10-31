/*
 * Copyright 2019 Splunk, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"): you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

package main

// Splunk Cloud Platform CLI
//
// Usage
//
//   ./scloud [options] <command> ...

//go:generate go run gen/gen_version.go

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"

	"github.com/golang/glog"
	"github.com/splunk/splunk-cloud-sdk-go/cmd/scloud/cli/assets"
	"github.com/splunk/splunk-cloud-sdk-go/cmd/scloud/cli/config"
	"github.com/splunk/splunk-cloud-sdk-go/cmd/scloud/cli/fcache"

	"github.com/splunk/splunk-cloud-sdk-go/cmd/scloud/cli/version"
	"golang.org/x/crypto/ssh/terminal"

	"strconv"

	"github.com/splunk/splunk-cloud-sdk-go/idp"
)

var options struct {
	env      string
	tenant   string
	username string
	password string
	noPrompt bool   // disable prompting for args
	authURL  string // scheme://host:port
	hostURL  string // scheme://host:port
	port     string
	insecure string // needs to be a string so we can test if the flag is set
	scheme   string
	certFile string
}

const SCloudHome = "SCLOUD_HOME"

var ctxCache *fcache.Cache
var settings *fcache.Cache

type multiFlags []string

func (i *multiFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *multiFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Returns an absolute path. If the given path is not absolute it looks
// for the environment variable SCLOUD HOME and is joined with that.
// If the environment variable is not defined, the path is joined with
// the path to the home dir
func abspath(p string) string {
	if path.IsAbs(p) {
		return p
	}
	scloudHome, ok := os.LookupEnv(SCloudHome)
	var root string
	var err error
	if ok {
		root = scloudHome
	} else {
		root, err = homedir.Dir()
		if err != nil {
			fatal(err.Error())
		}
	}
	return path.Join(root, p)
}

// Returns the name of the selected environment.
func getEnvironmentName() string {
	if options.env != "" {
		return options.env
	}
	if envName, ok := settings.GetString("env"); ok {
		return envName
	}
	if options.noPrompt {
		fatal("no environment")
	}
	envName := "prod" // default
	options.env = envName
	return envName
}

func getEnvironment() *config.Environment {
	name := getEnvironmentName()
	env, err := config.GetEnvironment(name)
	if err != nil {
		fatal(err.Error())
	}
	return env
}

// Returns the selected username.
func getUsername() string {
	if options.username != "" {
		return options.username
	}
	if username, ok := settings.GetString("username"); ok {
		return username
	}
	if options.noPrompt {
		fatal("no username")
	}
	var username string
	fmt.Print("Username: ")
	if _, err := fmt.Scanln(&username); err != nil {
		fatal(err.Error())
	}
	options.username = username
	return username
}

func getpass() (string, error) {
	fmt.Print("Password: ")
	data, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Returns the selected password.
func getPassword() string {
	if options.password != "" {
		return options.password
	}
	if options.noPrompt {
		fatal("no password")
	}
	password, err := getpass()
	if err != nil {
		fatal(err.Error())
	}
	return password
}

// Returns the selected app profile.
func getProfile() (map[string]string, error) {
	name := getProfileName()
	profile, err := config.GetProfile(name)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// Returns the name of the selected app profile.
func getProfileName() string {
	return getEnvironment().Profile
}

// Returns the selected tenant name.
func getTenantName() string {
	if options.tenant != "" {
		return options.tenant
	}
	if tenant, ok := settings.GetString("tenant"); ok {
		return tenant
	}
	if options.noPrompt {
		fatal("no tenant")
	}
	var tenant string
	fmt.Print("Tenant: ")
	if _, err := fmt.Scanln(&tenant); err != nil {
		fatal(err.Error())
	}
	options.tenant = tenant
	return tenant
}

// Returns auth url from passed-in options or local settings.
// If auth_url is not specified, returns ""
func getAuthURL() string {
	return getOptionSettings(options.authURL, "auth-url")
}

// Returns host url from passed-in options or local settings.
// If host_url is not specified, returns ""
func getHostURL() string {
	return getOptionSettings(options.hostURL, "host-url")
}

// Returns scheme from passed-in options or local settings.
// If ca-cert is not specified, returns ""
func getCaCert() string {
	return getOptionSettings(options.certFile, "ca-cert")
}

// Check the flag options first, fall back on settings
func getOptionSettings(option string, setting string) string {
	if option != "" {
		return option
	}
	if setting, ok := settings.GetString(setting); ok {
		return setting
	}
	return ""
}

// Defaults to false, reads from settings first.
// Overridden by --insecure flag
func isInsecure() bool {
	insecure := false
	var err error
	// local settings cache default value
	if insecureStr, ok := settings.GetString("insecure"); ok {
		insecure, err = strconv.ParseBool(insecureStr)
		if err != nil {
			insecure = false
		}
	}
	// --insecure=true passed as global flag
	if options.insecure != "" {
		insecure, err = strconv.ParseBool(options.insecure)
		if err != nil {
			insecure = false
		}
	}
	if insecure {
		glog.Warningf("TLS certificate validation is disabled.")
	}
	return insecure
}

// Prints an error message to stderr.
func eprint(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
}

// Prints error message and usage screen, then exits.
func eusage(msg string, args ...interface{}) {
	eprint(msg, args...)
	usage()
	os.Exit(1)
}

func etoofew() {
	fatal("too few arguments")
}

// Prints an error message and exits.
func fatal(msg string, args ...interface{}) {
	eprint(msg, args...)
	os.Exit(1)
}

func parseArgs() []string {
	flag.StringVar(&options.env, "env", "", "environment name")
	flag.StringVar(&options.username, "u", "", "user name")
	flag.StringVar(&options.password, "p", "", "password")
	flag.StringVar(&options.tenant, "tenant", "", "tenant name")
	flag.StringVar(&options.authURL, "auth-url", "", "auth url")
	flag.StringVar(&options.hostURL, "host-url", "", "host url")
	flag.BoolVar(&options.noPrompt, "no-prompt", false, "disable prompting")
	flag.StringVar(&options.insecure, "insecure", "", "disable tls cert validation")
	flag.StringVar(&options.certFile, "ca-cert", "", "client certificate file")
	flag.Parse()
	return flag.Args()
}

// Verify that the given list is empty, or fatal.
func checkEmpty(items []string) {
	if len(items) > 0 {
		fatal("unexpected arguments: '%s'", strings.Join(items, ", "))
	}
}

// [items] => head, [tail]
func head(items []string) (string, []string) {
	if items == nil {
		return "", nil
	}
	n := len(items)
	if n == 0 {
		return "", nil
	}
	h := items[0]
	if n == 1 {
		return h, nil
	}
	return h, items[1:]
}

func head1(items []string) string {
	if len(items) < 1 {
		etoofew()
	}
	h, items := head(items)
	checkEmpty(items)
	return h
}

func head2(items []string) (string, string) {
	if len(items) < 2 {
		etoofew()
	}
	h1, items := head(items)
	h2, items := head(items)
	checkEmpty(items)
	return h1, h2
}

// Returns head of vector withougt removing.
func peek(items []string) string { //nolint:deadcode
	if items == nil {
		return ""
	}
	n := len(items)
	if n == 0 {
		return ""
	}
	return items[0]
}

// "unget" the given argument by pushing back onto front of vector.
func push(item string, items []string) []string {
	return append([]string{item}, items...)
}

func pprint(value interface{}) {
	if value == nil {
		return
	}
	switch vt := value.(type) {
	case string:
		fmt.Print(vt)
		if !strings.HasSuffix(vt, "\n") {
			fmt.Println()
		}
	default:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "    ")
		err := encoder.Encode(value)
		if err != nil {
			fatal("json pprint error: %s", err.Error())
		}
	}
}

// Ensure that the given app profile contains the required user credentials.
func ensureCredentials(profile map[string]string) {
	kind, ok := profile["kind"]
	if !ok {
		return
	}
	if kind == "client" {
		return // user creds not needed
	}
	if _, ok := profile["username"]; !ok {
		profile["username"] = getUsername()
	}
	if _, ok := profile["password"]; !ok {
		profile["password"] = getPassword()
	}
}

// Returns the cached authorization context associated with the given clientID.
func getCurrentContext(clientID string) *idp.Context {
	v := ctxCache.Get(clientID)
	m, ok := v.(*toml.Tree)
	if !ok {
		glog.Warningf("Deleting context cache")
		// bad cache entry
		ctxCache.Delete(clientID) //nolint:errcheck
		return nil
	}
	context := &idp.Context{}
	if err := FromToml(context, m); err != nil {
		eprint(err.Error())
		// bad cache entry
		ctxCache.Delete(clientID) //nolint:errcheck
		return nil
	}
	return context
}

// Returns an authorization "context", which consists of the OAuth token(s)
// and related metadata that correspond to a given app. If a valid cached
// context exists, return those, otherwise dispatch an authn flow that
// corresponds to the selected app profile.
func getContext() *idp.Context {
	profile, err := getProfile()
	if err != nil {
		fatal(err.Error())
		return nil
	}
	clientID, ok := profile["client_id"]
	if !ok {
		fatal("bad app profile: no client_id")
		return nil
	}
	context := getCurrentContext(clientID)
	if context != nil {
		// todo: re-authenticate if token has expired
		return context
	}
	ensureCredentials(profile)
	context, err = authenticate(profile)
	if err != nil {
		fatal(err.Error())
		return nil
	}
	ctxCache.Set(clientID, Map(context))
	return context
}

func getToken() string {
	return getContext().AccessToken
}

// Authenticate, using the selected app profile.
func login(args []string) (*idp.Context, error) {
	checkEmpty(args)
	name := getProfileName()
	profile, err := config.GetProfile(name)
	if err != nil {
		return nil, err
	}
	clientID := profile["client_id"]
	glog.Infof("Authenticate profile=%s clientID=%s", name, clientID)
	ensureCredentials(profile)
	context, err := authenticate(profile)
	if err != nil {
		return nil, err
	}
	ctxCache.Set(clientID, Map(context))
	return context, nil
}

// Load config and settings.
func load() error {
	if err := loadConfig(); err != nil {
		return err
	}
	settings, _ = fcache.Load(abspath(".scloud"))
	ctxCache, _ = fcache.Load(abspath(".scloud_context"))
	return nil
}

// Load default config asset.
func loadConfig() error {
	file, err := assets.Open("default.yaml")
	if err != nil {
		return fmt.Errorf("err loading default.yaml: %s", err)
	}
	return config.Load(file)
}

// Display help text to stdout.
func callForHelp(args []string) bool {
	if len(args) >= 2 && args[1] == "help" {
		fileName := fmt.Sprintf("%s.txt", args[0])
		result, err := getHelp(fileName)
		if err != nil {
			fatal("%v", err)
		}
		fmt.Println(result)
		return true
	}
	return false
}

func main() {
	var err error

	args := parseArgs()
	if callForHelp(args) {
		return
	}

	buildTime := time.Unix(version.BuildTime, 0)
	glog.Infof("Version %s (%s) %s",
		version.Version,
		version.BuildBranch,
		buildTime.Format("2006-01-02 15:04:05"))

	if len(args) == 0 {
		eusage("missing command")
	}

	glog.CopyStandardLogTo("INFO")

	if err = load(); err != nil {
		fatal("internal error: %s", err.Error())
	}

	arg, args := head(args)

	var result interface{}
	switch arg {
	// authenticate using the selected oauth profile
	case "login":
		result, err = login(args)

	// delete the named setting
	case "delete":
		key := head1(args)
		err := settings.Delete(key)
		if err != nil {
			fatal("error deleting setting: %s", err.Error())
		}
		return

	// get the named setting
	case "get":
		key := head1(args)
		result = settings.Get(key)

	// set the value of the named setting
	case "set":
		key, value := head2(args)
		settings.Set(key, value)
		return

	// delete all settings
	case "delete-settings":
		checkEmpty(args)
		settings.Clear()
		return

	// get all settings
	case "get-settings":
		checkEmpty(args)
		result = settings.All()

	// reset the context cache
	case "delete-context-cache":
		checkEmpty(args)
		ctxCache.Clear()
		return

	// print context cache
	case "get-context-cache":
		checkEmpty(args)
		result = getContext()

	case "help":
		checkEmpty(args)
		result, err := getHelp("scloud.txt")
		if err == nil {
			fmt.Println(result)
		}

	case "action":
		result, err = newActionCommand().Dispatch(args)

	case "appreg":
		result, err = newAppRegistryCommand().Dispatch(args)

	case "catalog":
		result, err = newCatalogCommand().Dispatch(args)

	case "forwarders":
		result, err = newForwardersCommand(apiClient()).Dispatch(args)

	case "identity":
		result, err = newIdentityCommand().Dispatch(args)

	case "ingest":
		result, err = newIngestCommand().Dispatch(args)

	case "kvstore":
		result, err = newKVStoreCommand(apiClient()).Dispatch(args) // TODO: can share one apiClient() for all commands

	case "ml":
		result, err = newMachineLearningCommand().Dispatch(args)

	case "provisioner":
		result, err = newProvisionerCommand(apiClient()).Dispatch(args)

	case "search":
		result, err = newSearchCommand(apiClient()).Dispatch(args)

	case "streams":
		result, err = newStreamsCommand().Dispatch(args)

	case "version":
		fmt.Printf("scloud version %s-%s\n", version.Version, version.Commit)

	default:
		eusage("unknown command: '%s'", arg)
	}

	if err != nil {
		fatal("%v", err)
	}
	pprint(result)
}

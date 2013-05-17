/*
Internal workhorse of ginta. This package contains
few functions of interest to the user, but isolates the
core from the rest of the system to avoid accidental
misuse. 

It implements an actor that moderates all access to 
language resources. 
*/
package internal

import (
	types "code.google.com/p/ginta/common"
)

type fetchFunc func(string) <-chan types.Resource

type targetResource struct {
	target string
	types.Resource
}

type reply struct {
	tr  string
	err error
}

type request struct {
	key, code string
	recurse   bool
	reply     chan<- reply
}

type bundleRequest struct {
	recursive          bool
	code, bundlePrefix string
	reply              chan<- map[string]string
}

type languageRegister struct {
	key, name string
	fetch     fetchFunc
	c         chan bool
}

type languageActivate struct {
	code string
	done chan bool
}

type bundle map[string]string

type translation struct {
	displayName    string
	pendingFetches []fetchFunc
	runningFetches int
	blocked        []chan bool
	entries        map[string]bundle
}

var (
	requests         = make(chan request)
	bundleRequests   = make(chan bundleRequest)
	resourceEntry    = make(chan targetResource)
	fetchFinished    = make(chan string)
	registerLanguage = make(chan languageRegister)
	list             = make(chan chan<- []*types.Language)
	activateLanguage = make(chan languageActivate)

	universe = make(map[string]*translation)

	pending = 0
)

func init() {
	go work()
}

// Request a resource for a country code, either plain or recursively
func Request(code, key string, recurse bool) (string, error) {
	reply := make(chan reply)
	defer close(reply)

	requests <- request{key, code, recurse, reply}

	replyVal := <-reply
	return replyVal.tr, replyVal.err
}

// Requests a bundle for a prefix, either plain or recursively
func RequestBundle(code, base string, recursive bool) map[string]string {
	reply := make(chan map[string]string)
	defer close(reply)

	bundleRequests <- bundleRequest{
		code:         code,
		bundlePrefix: base,
		recursive:    recursive,
		reply:        reply,
	}

	return <-reply
}

// Registers a new language provider, using both a language and a resource enumerator function 
func Register(lang <-chan types.Language, fetch func(string) <-chan types.Resource) {
	for l := range lang {
		register := languageRegister{l.Code, l.DisplayName, fetchFunc(fetch), make(chan bool)}
		defer close(register.c)

		registerLanguage <- register
		<-register.c
	}
}

// Changes a mapped resource value
func Update(code, key, val string) {
	resourceEntry <- targetResource{code, types.Resource{key, val}}
}

// Lists all available languages
func List() []*types.Language {
	result := make(chan []*types.Language)
	defer close(result)

	list <- result

	return <-result
}

// makes a language ready for use by loading all associated resources
func Activate(code string) bool {
	c := make(chan bool)
	defer close(c)
	activateLanguage <- languageActivate{code, c}

	return <-c
}

func work() {
	for {
		select {
		case l := <-registerLanguage:
			doRegister(&l)
		case activate := <-activateLanguage:
			doActivate(&activate)
		case finished := <-fetchFinished:
			doFetchFinished(finished)
		case entry := <-resourceEntry:
			doAddEntry(&entry)
		case request := <-bundleRequests:
			result := make(map[string]string)
			mergeBundles(result, types.HierarchicalKey(request.bundlePrefix), &request)

			request.reply <- result
		case request := <-requests:
			doFetchResource(&request)
		case reply := <-list:
			doListLanguages(reply)
		}
	}
}

func mergeBundles(result map[string]string, k types.HierarchicalKey, request *bundleRequest) {
	if entries, ok := universe[request.code].entries[k.String()]; ok {
		for key, val := range entries {

			if _, ok := result[key]; !ok {
				result[key] = val
			}
		}
	}

	if request.recursive && k.String() != "" {
		mergeBundles(result, k.Parent(), request)
	}
}

func doActivate(activate *languageActivate) {
	if entry, ok := universe[activate.code]; ok {
		if length := len(entry.pendingFetches); length > 0 {
			entry.runningFetches += length
			entry.blocked = append(entry.blocked, activate.done)
			list := entry.pendingFetches
			entry.pendingFetches = []fetchFunc{}
			for _, fun := range list {
				code := activate.code
				fun := fun
				go func() {
					for next := range fun(activate.code) {

						resourceEntry <- targetResource{code, next}
					}

					fetchFinished <- code
				}()
			}
		} else {
			activate.done <- true
		}
	} else {
		activate.done <- false
	}
}

func doRegister(l *languageRegister) {
	entry, ok := universe[l.key]
	if !ok {
		entry = &translation{
			displayName:    l.name,
			pendingFetches: []fetchFunc{},
			blocked:        []chan bool{},
			entries:        make(map[string]bundle),
		}
		universe[l.key] = entry
	}

	entry.pendingFetches = append(entry.pendingFetches, l.fetch)

	l.c <- !ok
}

func doFetchFinished(finished string) {
	entry := universe[finished]
	entry.runningFetches--

	if entry.runningFetches == 0 {
		for _, c := range entry.blocked {
			c <- true
		}

		entry.blocked = make([]chan bool, 0)
	}
}

func doAddEntry(entry *targetResource) {
	if ptr := universe[entry.target]; ptr != nil {
		prefix, key := types.HierarchicalKey(entry.Key).Split()
		m, ok := ptr.entries[prefix]
		if !ok {
			m = make(bundle)
			ptr.entries[prefix] = m
		}
		m[key] = entry.Value
	}
}

func doFetchResource(request *request) {
	if lang, ok := universe[request.code]; ok {

		iteration := true

		hierarchy := types.HierarchicalKey(request.key)

		for iteration {
			prefix, key := hierarchy.Split()

			if m, ok := lang.entries[prefix]; ok {
				if str, ok := m[key]; ok {
					request.reply <- reply{str, nil}
					return
				}
			}

			hierarchy = hierarchy.Parent()
			iteration = request.recurse && hierarchy.String() != ""
		}
	}

	request.reply <- reply{request.key, types.ResourceNotFoundError(request.key)}
}

func doListLanguages(reply chan<- []*types.Language) {
	result := make([]*types.Language, len(universe))
	i := 0
	for key, val := range universe {
		result[i] = &types.Language{key, val.displayName}
		i++
	}

	reply <- result
}

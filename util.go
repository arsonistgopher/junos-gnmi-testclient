package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	stdpath "path"
	"sort"
	"strings"

	gnmipb "github.com/arsonistgopher/junos-gnmi-testclient/proto/gnmi"
)

func getSMode(mode string) gnmipb.SubscriptionMode {
	switch strings.ToLower(mode) {
	case "target-defined":
		return gnmipb.SubscriptionMode_TARGET_DEFINED
	case "sample":
		return gnmipb.SubscriptionMode_SAMPLE
	case "on-change":
		return gnmipb.SubscriptionMode_ON_CHANGE
	default:
		log.Fatalf("unsupported subscription mode\n")
	}
	return gnmipb.SubscriptionMode_SAMPLE
}

func getMode(mode string) gnmipb.SubscriptionList_Mode {
	switch strings.ToLower(mode) {
	case "once":
		return gnmipb.SubscriptionList_ONCE
	case "stream":
		return gnmipb.SubscriptionList_STREAM
	case "poll":
		return gnmipb.SubscriptionList_POLL
	default:
		log.Fatalf("unsupported mode: please use stream | once | poll \n")
	}
	return gnmipb.SubscriptionList_STREAM
}

func pathToString(q []string) string {
	qq := make([]string, len(q))
	copy(qq, q)
	for i, e := range qq {
		qq[i] = strings.Replace(e, "/", "\\/", -1)
	}
	return strings.Join(qq, "/")
}

func xpathToGNMIpath(input string) ([]string, error) {
	path := strings.Trim(input, "/")
	var buf []rune
	inKey := false
	null := rune(0)
	for _, r := range path {
		switch r {
		case '[':
			if inKey {
				return nil, fmt.Errorf("malformed path, nested '[': %q ", path)
			}
			inKey = true
		case ']':
			if !inKey {
				return nil, fmt.Errorf("malformed path, unmatched ']': %q", path)
			}
			inKey = false
		case '/':
			if !inKey {
				buf = append(buf, null)
				continue
			}
		}
		buf = append(buf, r)
	}
	if inKey {
		return nil, fmt.Errorf("malformed path, missing trailing ']': %q", path)
	}
	return strings.Split(string(buf), string(null)), nil
}

// PathType export
type PathType int64

const (
	// StructuredPath export
	StructuredPath PathType = iota
	// StringSlicePath export
	StringSlicePath
)

// PathToString export
func PathToString(path *gnmipb.Path) (string, error) {
	s, err := PathToStrings(path)
	return "/" + stdpath.Join(s...), err
}

// PathToStrings export
func PathToStrings(path *gnmipb.Path) ([]string, error) {
	var p []string
	if path.Element != nil {
		for i, e := range path.Element {
			if e == "" {
				return nil, fmt.Errorf("empty element at index %d in %v", i, path.Element)
			}
			p = append(p, e)
		}
		return p, nil
	}

	for i, e := range path.Elem {
		if e.Name == "" {
			return nil, fmt.Errorf("empty name for PathElem at index %d", i)
		}

		elem, err := elemToString(e.Name, e.Key)
		if err != nil {
			return nil, fmt.Errorf("failed formatting PathElem at index %d: %v", i, err)
		}
		p = append(p, elem)
	}
	return p, nil
}

func elemToString(name string, kv map[string]string) (string, error) {
	if name == "" {
		return "", errors.New("empty name for PathElem")
	}
	if len(kv) == 0 {
		return name, nil
	}

	var keys []string
	for k, v := range kv {
		if k == "" {
			return "", fmt.Errorf("empty key name (value: %s) in element %s", v, name)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := strings.Replace(kv[k], `=`, `\=`, -1)
		v = strings.Replace(v, `]`, `\]`, -1)
		name = fmt.Sprintf("%s[%s=%s]", name, k, v)
	}

	return name, nil
}

// StringToPath export
func StringToPath(path string, pathTypes ...PathType) (*gnmipb.Path, error) {
	var errs Errors
	if len(pathTypes) == 0 {
		return nil, AppendErr(errs, errors.New("no path types specified"))
	}

	pmsg := &gnmipb.Path{}
	for _, p := range pathTypes {
		switch p {
		case StructuredPath:
			gp, err := StringToStructuredPath(path)
			if err != nil {
				errs = AppendErr(errs, fmt.Errorf("error building structured path: %v", err))
				continue
			}
			pmsg.Elem = gp.Elem
		case StringSlicePath:
			gp, err := StringToStringSlicePath(path)
			if err != nil {
				errs = AppendErr(errs, fmt.Errorf("error building string slice path: %v", err))
				continue
			}
			pmsg.Element = gp.Element
		}
	}

	if errs != nil {
		return nil, errs
	}

	return pmsg, nil
}

// StringToStringSlicePath export
func StringToStringSlicePath(path string) (*gnmipb.Path, error) {
	parts := pathStringToElements(path)
	gpath := new(gnmipb.Path)
	for _, p := range parts {
		name, kv, err := extractKV(p)
		if err != nil {
			return nil, fmt.Errorf("error parsing path %q: %v", path, err)
		}
		fpath, err := elemToString(name, kv)
		if err != nil {
			return nil, fmt.Errorf("error formatting path %q: %v", path, err)
		}
		gpath.Element = append(gpath.Element, fpath)
	}

	return gpath, nil
}

// StringToStructuredPath export
func StringToStructuredPath(path string) (*gnmipb.Path, error) {
	parts := pathStringToElements(path)

	gpath := &gnmipb.Path{}
	for _, p := range parts {
		name, kv, err := extractKV(p)
		if err != nil {
			return nil, fmt.Errorf("error parsing path %s: %v", path, err)
		}
		gpath.Elem = append(gpath.Elem, &gnmipb.PathElem{
			Name: name,
			Key:  kv,
		})
	}
	return gpath, nil
}

func pathStringToElements(s string) []string {
	var parts []string
	var buf bytes.Buffer

	var inKey, inEscape bool

	for _, ch := range s {
		switch {
		case ch == '[' && !inEscape:
			inKey = true
		case ch == ']' && !inEscape:
			inKey = false
		case ch == '\\' && !inEscape && !inKey:
			inEscape = true
			continue
		case ch == '/' && !inEscape && !inKey:
			parts = append(parts, buf.String())
			buf.Reset()
			continue
		}

		buf.WriteRune(ch)
		inEscape = false
	}

	if buf.Len() != 0 {
		parts = append(parts, buf.String())
	}

	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	return parts
}

func extractKV(in string) (string, map[string]string, error) {
	var inEscape, inKey, inValue bool
	var name, currentKey string
	var buf bytes.Buffer
	keys := map[string]string{}

	for _, ch := range in {
		switch {
		case ch == '[' && !inEscape && !inValue && inKey:
			return "", nil, fmt.Errorf("received an unescaped [ in key of element %s", name)
		case ch == '[' && !inEscape && !inKey:
			inKey = true
			if len(keys) == 0 {
				if buf.Len() == 0 {
					return "", nil, errors.New("received a value when the element name was null")
				}
				name = buf.String()
				buf.Reset()
			}
			continue
		case ch == ']' && !inEscape && !inKey:
			return "", nil, fmt.Errorf("received an unescaped ] when not in a key for element %s", buf.String())
		case ch == ']' && !inEscape:
			inKey = false
			inValue = false
			if err := addKey(keys, name, currentKey, buf.String()); err != nil {
				return "", nil, err
			}
			buf.Reset()
			currentKey = ""
			continue
		case ch == '\\' && !inEscape:
			inEscape = true
			continue
		case ch == '=' && inKey && !inEscape && !inValue:
			currentKey = buf.String()
			buf.Reset()
			inValue = true
			continue
		}

		buf.WriteRune(ch)
		inEscape = false
	}

	if len(keys) == 0 {
		name = buf.String()
	}

	if len(keys) != 0 && buf.Len() != 0 {
		return "", nil, fmt.Errorf("trailing garbage following keys in element %s, got: %v", name, buf.String())
	}

	if strings.Contains(name, " ") {
		return "", nil, fmt.Errorf("invalid space character included in element name '%s'", name)
	}

	return name, keys, nil
}

func addKey(keys map[string]string, e, k, v string) error {
	switch {
	case strings.Contains(k, " "):
		return fmt.Errorf("received an invalid space in element %s key name '%s'", e, k)
	case e == "":
		return fmt.Errorf("received null element value with key and value %s=%s", k, v)
	case k == "":
		return fmt.Errorf("received null key name for element %s", e)
	case v == "":
		return fmt.Errorf("received null value for key %s of element %s", k, e)
	}
	keys[k] = v
	return nil
}

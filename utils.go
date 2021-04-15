package ffmpeg_go

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"

	"github.com/u2takey/go-utils/sets"
)

func getString(item interface{}) string {
	if a, ok := item.(interface{ String() string }); ok {
		return a.String()
	}
	switch a := item.(type) {
	case string:
		return a
	case []string:
		return strings.Join(a, ", ")
	case Args:
		return strings.Join(a, ", ")
	case []interface{}:
		var r []string
		for _, b := range a {
			r = append(r, getString(b))
		}
		return strings.Join(r, ", ")
	case KwArgs:
		var keys, r []string
		for k := range a {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			r = append(r, fmt.Sprintf("%s: %s", k, getString(a[k])))
		}
		return fmt.Sprintf("{%s}", strings.Join(r, ", "))
	case map[string]interface{}:
		var keys, r []string
		for k := range a {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			r = append(r, fmt.Sprintf("%s: %s", k, getString(a[k])))
		}
		return fmt.Sprintf("{%s}", strings.Join(r, ", "))
	}
	return fmt.Sprintf("%v", item)
}

func getHash(item interface{}) int {
	h := fnv.New64()
	switch a := item.(type) {
	case interface{ Hash() int }:
		return a.Hash()
	case string:
		_, _ = h.Write([]byte(a))
		return int(h.Sum64())
	case []byte:
		_, _ = h.Write(a)
		return int(h.Sum64())
	case map[string]interface{}:
		b := 0
		for k, v := range a {
			b += getHash(k) + getHash(v)
		}
		return b
	case KwArgs:
		b := 0
		for k, v := range a {
			b += getHash(k) + getHash(v)
		}
		return b
	default:
		_, _ = h.Write([]byte(getString(item)))
		return int(h.Sum64())
	}
}

func escapeChars(text, chars string) string {
	s := sets.NewString()
	for _, a := range chars {
		s.Insert(string(a))
	}
	sl := s.List()
	if s.Has("\\") {
		s.Delete("\\")
		sl = append([]string{"\\"}, s.List()...)
	}
	for _, ch := range sl {
		text = strings.ReplaceAll(text, ch, "\\"+ch)
	}
	return text
}

type Args []string

func (a Args) Sorted() Args {
	sort.Strings(a)
	return a
}

func (a Args) EscapeWith(chars string) Args {
	out := Args{}
	for _, b := range a {
		out = append(out, escapeChars(b, chars))
	}
	return out
}

type KwArgs map[string]interface{}

func MergeKwArgs(args []KwArgs) KwArgs {
	a := KwArgs{}
	for _, b := range args {
		for c := range b {
			a[c] = b[c]
		}
	}
	return a
}

func (a KwArgs) EscapeWith(chars string) KwArgs {
	out := KwArgs{}
	for k, v := range a {
		out[escapeChars(k, chars)] = escapeChars(getString(v), chars)
	}
	return out
}

func (a KwArgs) Copy() KwArgs {
	r := KwArgs{}
	for k := range a {
		r[k] = a[k]
	}
	return r
}

func (a KwArgs) SortedKeys() []string {
	var r []string
	for k := range a {
		r = append(r, k)
	}
	sort.Strings(r)
	return r
}

func (a KwArgs) GetString(k string) string {
	if v, ok := a[k]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func (a KwArgs) PopString(k string) string {
	if c, ok := a[k]; ok {
		defer delete(a, k)
		return fmt.Sprintf("%v", c)
	}
	return ""
}

func (a KwArgs) HasKey(k string) bool {
	_, ok := a[k]
	return ok
}

func (a KwArgs) GetDefault(k string, defaultV interface{}) interface{} {
	if v, ok := a[k]; ok {
		return v
	}
	return defaultV
}

func (a KwArgs) PopDefault(k string, defaultV interface{}) interface{} {
	if v, ok := a[k]; ok {
		defer delete(a, k)
		return v
	}
	return defaultV
}

func ConvertKwargsToCmdLineArgs(kwargs KwArgs) []string {
	var keys, args []string
	for k := range kwargs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := kwargs[k]
		switch a := v.(type) {
		case string:
			args = append(args, fmt.Sprintf("-%s", k))
			if a != "" {
				args = append(args, a)
			}
		case []string:
			for _, r := range a {
				args = append(args, fmt.Sprintf("-%s", k))
				if r != "" {
					args = append(args, r)
				}
			}
		case []int:
			for _, r := range a {
				args = append(args, fmt.Sprintf("-%s", k))
				args = append(args, strconv.Itoa(r))
			}
		case int:
			args = append(args, fmt.Sprintf("-%s", k))
			args = append(args, strconv.Itoa(a))
		default:
			args = append(args, fmt.Sprintf("-%s", k))
			args = append(args, fmt.Sprintf("%v", a))
		}
	}
	return args
}

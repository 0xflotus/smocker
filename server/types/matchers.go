package types

import (
	"encoding/json"
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/smartystreets/assertions"
	"github.com/stretchr/objx"
	"gopkg.in/yaml.v3"
)

const (
	DefaultMatcher = "ShouldEqual"
)

type Assertion func(actual interface{}, expected ...interface{}) string

var asserts = map[string]Assertion{
	"ShouldResemble":         assertions.ShouldResemble,
	"ShouldAlmostEqual":      assertions.ShouldAlmostEqual,
	"ShouldContainSubstring": assertions.ShouldContainSubstring,
	"ShouldEndWith":          assertions.ShouldEndWith,
	"ShouldEqual":            assertions.ShouldEqual,
	"ShouldEqualJSON":        assertions.ShouldEqualJSON,
	"ShouldStartWith":        assertions.ShouldStartWith,
	"ShouldBeEmpty":          ShouldBeEmpty,
	"ShouldMatch":            ShouldMatch,

	"ShouldNotResemble":         assertions.ShouldNotResemble,
	"ShouldNotAlmostEqual":      assertions.ShouldNotAlmostEqual,
	"ShouldNotContainSubstring": assertions.ShouldNotContainSubstring,
	"ShouldNotEndWith":          assertions.ShouldNotEndWith,
	"ShouldNotEqual":            assertions.ShouldNotEqual,
	"ShouldNotStartWith":        assertions.ShouldNotStartWith,
	"ShouldNotBeEmpty":          ShouldNotBeEmpty,
	"ShouldNotMatch":            ShouldNotMatch,
}

func ShouldMatch(value interface{}, patterns ...interface{}) string {
	valueString, ok := value.(string)
	if !ok {
		return "ShouldMatch works only with strings"
	}

	for _, pattern := range patterns {
		patternString, ok := pattern.(string)
		if !ok {
			return "ShouldMatch works only with strings"
		}

		if match, err := regexp.MatchString(patternString, valueString); !match || err != nil {
			return fmt.Sprintf("Expected %q to match %q (but it didn't)!", valueString, patternString)
		}
	}

	return ""
}

func ShouldBeEmpty(value interface{}, patterns ...interface{}) string {
	return assertions.ShouldBeEmpty(value)
}

func ShouldNotBeEmpty(value interface{}, patterns ...interface{}) string {
	return assertions.ShouldNotBeEmpty(value)
}

func ShouldNotMatch(value interface{}, patterns ...interface{}) string {
	valueString, ok := value.(string)
	if !ok {
		return "ShouldNotMatch works only with strings"
	}

	for _, pattern := range patterns {
		patternString, ok := pattern.(string)
		if !ok {
			return "ShouldNotMatch works only with strings"
		}

		if match, err := regexp.MatchString(patternString, valueString); match && err == nil {
			return fmt.Sprintf("Expected %q to not match %q (but it did)!", valueString, patternString)
		}
	}

	return ""
}

type StringMatcher struct {
	Matcher string
	Value   string
}

func (sm StringMatcher) Match(value string) bool {
	matcher := asserts[sm.Matcher]
	if matcher == nil {
		log.WithField("matcher", sm.Matcher).Error("Invalid matcher")
		return false
	}

	if res := matcher(value, sm.Value); res != "" {
		log.Tracef("Value doesn't match:\n%s", res)
		return false
	}

	return true
}

func (sm StringMatcher) MarshalJSON() ([]byte, error) {
	if sm.Matcher == DefaultMatcher {
		return json.Marshal(sm.Value)
	}

	return json.Marshal(&struct {
		Matcher string `json:"matcher"`
		Value   string `json:"value"`
	}{
		Matcher: sm.Matcher,
		Value:   sm.Value,
	})
}

func (sm *StringMatcher) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		sm.Matcher = DefaultMatcher
		sm.Value = s
		return nil
	}

	var res struct {
		Matcher string `json:"matcher"`
		Value   string `json:"value"`
	}

	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	sm.Matcher = res.Matcher
	sm.Value = res.Value
	return nil
}

func (sm StringMatcher) MarshalYAML() (interface{}, error) {
	if sm.Matcher == DefaultMatcher {
		return sm.Value, nil
	}

	value, err := yaml.Marshal(&struct {
		Matcher string `yaml:"matcher,flow"`
		Value   string `yaml:"value,flow"`
	}{
		Matcher: sm.Matcher,
		Value:   sm.Value,
	})

	return string(value), err
}

func (sm *StringMatcher) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		sm.Matcher = DefaultMatcher
		sm.Value = s
		return nil
	}

	var res struct {
		Matcher string `yaml:"matcher,flow"`
		Value   string `yaml:"value,flow"`
	}

	if err := unmarshal(&res); err != nil {
		return err
	}

	sm.Matcher = res.Matcher
	sm.Value = res.Value
	return nil
}

type StringMatcherSlice []StringMatcher

func (sms StringMatcherSlice) Match(values []string) bool {
	if len(sms) > len(values) {
		return false
	}
	for _, matcher := range sms {
		matched := false
		for _, v := range values {
			if matcher.Match(v) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (sms StringMatcherSlice) MarshalJSON() ([]byte, error) {
	if len(sms) == 1 {
		if sms[0].Matcher == DefaultMatcher {
			return json.Marshal(sms[0].Value)
		}
		return json.Marshal(sms[0])
	}
	res := make([]StringMatcher, len(sms))
	for i, v := range sms {
		res[i] = StringMatcher{
			Matcher: v.Matcher,
			Value:   v.Value,
		}
	}
	return json.Marshal(res)
}

func (sms *StringMatcherSlice) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*sms = []StringMatcher{{
			Matcher: DefaultMatcher,
			Value:   s,
		}}
		return nil
	}

	var sm StringMatcher
	if err := json.Unmarshal(data, &sm); err == nil {
		*sms = []StringMatcher{sm}
		return nil
	}

	var res []StringMatcher
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	*sms = res
	return nil
}

func (sms StringMatcherSlice) MarshalYAML() (interface{}, error) {
	if len(sms) == 1 {
		if sms[0].Matcher == DefaultMatcher {
			value, err := yaml.Marshal(sms[0].Value)
			return string(value), err
		}
		value, err := yaml.Marshal(sms[0])
		return string(value), err
	}
	res := make([]StringMatcher, len(sms))
	for i, v := range sms {
		res[i] = StringMatcher{
			Matcher: v.Matcher,
			Value:   v.Value,
		}
	}
	value, err := yaml.Marshal(res)
	return string(value), err
}

func (sms *StringMatcherSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		*sms = []StringMatcher{{
			Matcher: DefaultMatcher,
			Value:   s,
		}}
		return nil
	}

	var sm StringMatcher
	if err := unmarshal(&sm); err == nil {
		*sms = []StringMatcher{sm}
		return nil
	}

	var res []StringMatcher
	if err := unmarshal(&res); err != nil {
		return err
	}
	*sms = res
	return nil
}

type MultiMapMatcher map[string]StringMatcherSlice

func (mmm MultiMapMatcher) Match(values map[string][]string) bool {
	if len(mmm) > len(values) {
		return false
	}
	for key, matcherValue := range mmm {
		value, ok := values[key]
		if !ok || !matcherValue.Match(value) {
			return false
		}
	}
	return true
}

type BodyMatcher struct {
	bodyString *StringMatcher
	bodyJson   map[string]StringMatcher
}

func (bm BodyMatcher) Match(value string) bool {
	if bm.bodyString != nil {
		return bm.bodyString.Match(value)
	}

	j, err := objx.FromJSON(value)
	if err != nil {
		return false
	}
	for path, matcher := range bm.bodyJson {
		value := j.Get(path)
		if value == nil {
			return false
		}
		if ok := matcher.Match(value.String()); !ok {
			return false
		}
	}
	return true
}

func (bm BodyMatcher) MarshalJSON() ([]byte, error) {
	if bm.bodyString != nil {
		return json.Marshal(bm.bodyString)
	}
	return json.Marshal(bm.bodyJson)
}

func (bm *BodyMatcher) UnmarshalJSON(data []byte) error {
	var s StringMatcher
	if err := json.Unmarshal(data, &s); err == nil {
		if _, ok := asserts[s.Matcher]; ok {
			bm.bodyString = &s
			return nil
		}
	}

	var res map[string]StringMatcher
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	bm.bodyJson = res
	return nil
}

func (bm BodyMatcher) MarshalYAML() (interface{}, error) {
	if bm.bodyString != nil {
		value, err := yaml.Marshal(bm.bodyString)
		return string(value), err
	}

	value, err := yaml.Marshal(bm.bodyJson)
	return string(value), err
}

func (bm *BodyMatcher) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s StringMatcher
	if err := unmarshal(&s); err == nil {
		if _, ok := asserts[s.Matcher]; ok {
			bm.bodyString = &s
			return nil
		}
	}

	var res map[string]StringMatcher
	if err := unmarshal(&res); err != nil {
		return err
	}
	bm.bodyJson = res
	return nil
}

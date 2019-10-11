/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"errors"
	"reflect"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/grafeas/kritis/pkg/kritis/constants"
)

// RemoveGloballyAllowedImages returns all images that aren't in a global allowlist
func RemoveGloballyAllowedImages(images []string) []string {
	notAllowlisted := []string{}
	for _, image := range images {
		allowlisted, err := imageInGlobalAllowlist(image)
		if err != nil {
			glog.Errorf("couldn't check if %s is in global allowlist: %v", image, err)
		}
		if !allowlisted {
			notAllowlisted = append(notAllowlisted, image)
		}
	}
	return notAllowlisted
}

// Do an image match based on reference.
func imageRefMatch(image string, pattern string) (bool, error) {
	allowRef, err := name.ParseReference(pattern, name.WeakValidation)
	if err != nil {
		return false, err
	}
	imageRef, err := name.ParseReference(image, name.WeakValidation)
	if err != nil {
		return false, err
	}
	// Make sure images have the same context
	if reflect.DeepEqual(allowRef.Context(), imageRef.Context()) {
		return true, nil
	}
	return false, nil
}

// Do an image match based on name pattern.
// See https://cloud.google.com/binary-authorization/docs/policy-yaml-reference#admissionwhitelistpatterns
func imageNamePatternMatch(image string, pattern string) (bool, error) {
	if len(pattern) == 0 {
		return false, errors.New("empty pattern")
	}
	if pattern[len(pattern)-1] == '*' {
		pattern = pattern[:len(pattern)-1]
		if strings.HasPrefix(image, pattern) {
			if strings.LastIndex(image, "/") < len(pattern) {
				return true, nil
			}
		}
	} else {
		if image == pattern {
			return true, nil
		}
	}
	return false, nil
}

func imageInAllowlistByReference(image string, allowList []string) (bool, error) {
	for _, w := range allowList {
		match, err := imageRefMatch(image, w)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func imageInAllowlistByPattern(image string, allowList []string) (bool, error) {
	for _, w := range allowList {
		match, err := imageNamePatternMatch(image, w)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

// Check if image is allowed by global allowlist.
// This method uses reference matching.
func imageInGlobalAllowlist(image string) (bool, error) {
	return imageInAllowlistByReference(image, constants.GlobalImageAllowlist)
}

// Check if image is allowed by a GAP allowlist.
// This method uses name pattern matching.
func imageInGapAllowlist(image string, allowlist []string) (bool, error) {
	return imageInAllowlistByPattern(image, allowlist)
}

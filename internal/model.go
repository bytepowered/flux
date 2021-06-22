package internal

import (
	"fmt"
	"github.com/bytepowered/flux"
	"regexp"
)

var (
	annotationKeySpecPattern   = regexp.MustCompile(`^[a-z0-9A-Z]+[a-z0-9A-Z-_.\/]*`)
	annotationKeySpecMaxLength = 255
)

func VerifyAnnotationKeySpec(key string) error {
	if "" == key || len(key) > annotationKeySpecMaxLength {
		return fmt.Errorf("annotation key size invalid, key: %s", key)
	}
	if !annotationKeySpecPattern.MatchString(key) {
		return fmt.Errorf("annotation key spec invalid, key: %s", key)
	}
	return nil
}

func VerifyAnnotations(annotations flux.Annotations) error {
	for key := range annotations {
		err := VerifyAnnotationKeySpec(key)
		if err != nil {
			return err
		}
	}
	return nil
}

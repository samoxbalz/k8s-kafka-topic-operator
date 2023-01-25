package util

import (
	"github.com/mitchellh/hashstructure/v2"
	"log"
	"strconv"
)

const HashLabelName = "zetaedge.io/hash"

func GenerateHashFromSpec(spec interface{}) string {
	hash, err := hashstructure.Hash(spec, hashstructure.FormatV2, nil)
	if err != nil {
		log.Fatalln(err)
	}
	return strconv.FormatUint(hash, 10)
}

func SetHashToLabel(labels map[string]string, hashValue string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[HashLabelName] = hashValue
	return labels
}

func GetHashFromLabels(labels map[string]string) string {
	return labels[HashLabelName]
}

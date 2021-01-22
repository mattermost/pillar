// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package utils

import (
	"bytes"
	"encoding/base32"

	"github.com/pborman/uuid"
)

func GetDisplayableCardBrand(brand string) string {
	switch brand {
	case "amex":
		return "American Express"
	case "diners":
		return "Diner's Club"
	case "discover":
		return "Discover"
	case "jcb":
		return "JCB"
	case "mastercard":
		return "Mastercard"
	case "visa":
		return "VISA"
	}

	return brand
}

var encoding = base32.NewEncoding("ybndrfg8ejkmcpqxot1uwisza345h769")

// NewID is a globally unique identifier.  It is a [A-Z0-9] string 26
// characters long.  It is a UUID version 4 Guid that is zbased32 encoded
// with the padding stripped off.
func NewID() string {
	var b bytes.Buffer
	encoder := base32.NewEncoder(encoding, &b)
	_, _ = encoder.Write(uuid.NewRandom())
	encoder.Close()
	b.Truncate(26) // removes the '==' padding
	return b.String()
}

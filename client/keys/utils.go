// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)
package keys

import (
	"encoding/json"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cryptokeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// available output formats.
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
)

type bechKeyOutFn func(k *cryptokeyring.Record) (keys.KeyOutput, error)

func printKeyringRecord(w io.Writer, k *cryptokeyring.Record, bechKeyOut bechKeyOutFn, output string) error {
	ko, err := bechKeyOut(k)
	if err != nil {
		return err
	}

	switch output {
	case OutputFormatText:
		if err := printTextRecords(w, []keys.KeyOutput{ko}); err != nil {
			return err
		}

	case OutputFormatJSON:
		out, err := json.Marshal(ko)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return err
		}
	}

	return nil
}

func printTextRecords(w io.Writer, kos []keys.KeyOutput) error {
	out, err := yaml.Marshal(&kos)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, string(out)); err != nil {
		return err
	}

	return nil
}

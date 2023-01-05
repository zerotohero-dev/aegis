/*
 * .-'_.---._'-.
 * ||####|(__)||   Protect your secrets, protect your business.
 *   \\()|##//       Secure your sensitive data with Aegis.
 *    \\ |#//                  <aegis.z2h.dev>
 *     .\_/.
 */

package route

import (
	v1 "aegis-safe/internal/entity/reqres/v1"
	"aegis-safe/internal/state"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

func Secret(w http.ResponseWriter, r *http.Request, svid string) {
	if r == nil {
		return
	}

	// TODO: move these validations to a common module.
	if !strings.HasPrefix(svid, "spiffe://aegis.z2h.dev/workload/aegis-sentinel/ns/aegis-system/sa/aegis-sentinel/n/") {
		// TODO: return an error response instead.
		_, _ = io.WriteString(w, "")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		// TODO: return an error response instead.
		_, _ = io.WriteString(w, "")
	}
	defer func(b io.ReadCloser) {
		if b == nil {
			return
		}
		err := b.Close()
		if err != nil {
			log.Println("Problem closing body")
		}
	}(r.Body)

	var sr v1.SecretUpsertRequest

	err = json.Unmarshal(body, &sr)
	if err != nil {
		// TODO: return an error response instead.
		_, _ = io.WriteString(w, "")
		return
	}

	workloadId := sr.WorkloadId
	value := sr.Value

	state.UpsertSecret(workloadId, value)

	// TODO: handle these
	_, _ = io.WriteString(w, "OK")
}
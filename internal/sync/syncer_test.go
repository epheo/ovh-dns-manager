package sync

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
)

func TestPrintSummaryShowsErrorsWithoutChanges(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	(&SyncResult{Errors: []error{errors.New("failed to create record :MX: 400")}}).PrintSummary()
	if !strings.Contains(buf.String(), "failed to create record :MX: 400") {
		t.Fatalf("the only operation failed and the log does not say why:\n%s", buf.String())
	}
}

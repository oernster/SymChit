package application

import (
	"strings"
	"testing"

	"github.com/oernster/symchit/internal/product"
)

func TestAboutCarriesTheStatement(t *testing.T) {
	t.Parallel()
	about := NewAbout("1.2.3")
	if about.Name != product.Name || about.Version != "1.2.3" || about.Author != product.Author {
		t.Errorf("about = %+v", about)
	}
	if !strings.Contains(about.Statement, "no medical advice") {
		t.Errorf("the statement does not say it gives no medical advice: %q", about.Statement)
	}
	if len(about.Credits) == 0 {
		t.Error("no credits")
	}
	about.Credits[0].Work = "changed"
	if NewAbout("1.2.3").Credits[0].Work == "changed" {
		t.Error("a caller can change the shared credit list")
	}
}

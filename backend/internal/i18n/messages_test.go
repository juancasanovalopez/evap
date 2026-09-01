package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestText_UsesRegisteredTranslations(t *testing.T) {
	tests := []struct {
		name string
		tag  language.Tag
		want string
	}{
		{name: "English", tag: language.English, want: "date range cannot exceed 31 days"},
		{name: "Spanish", tag: language.Spanish, want: "el rango de fechas no puede superar 31 días"},
		{name: "French", tag: language.French, want: "la plage de dates ne peut pas dépasser 31 jours"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, Text(test.tag, SimulationRangeTooLarge, 31))
		})
	}
}

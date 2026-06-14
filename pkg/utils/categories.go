package utils

import "strings"

// PredefinedCategories is the canonical list of expense categories.
// Both Spanish and English names are valid, normalized to Spanish for storage.
var PredefinedCategories = []CategoryDef{
	{Name: "Comida", Aliases: []string{"food", "comidas", "restaurante", "restaurant", "delivery", "pedidosya", "rappipago"}},
	{Name: "Supermercado", Aliases: []string{"supermarket", "super", "mercado", "almacen", "almacén", "chino", "chino"}},
	{Name: "Vivienda", Aliases: []string{"housing", "alquiler", "rent", "casa", "departamento", "expensas"}},
	{Name: "Transporte", Aliases: []string{"transport", "transporte", "taxi", "uber", "didit", "colectivo", "bus", "subte", "nafta", "combustible"}},
	{Name: "Servicios", Aliases: []string{"services", "service", "luz", "gas", "agua", "internet", "telefono", "teléfono", "celular", "streaming", "netflix", "spotify"}},
	{Name: "Entretenimiento", Aliases: []string{"entertainment", "ocio", "salidas", "cine", "teatro", "show", "concierto", "bar", "cerveza"}},
	{Name: "Salud", Aliases: []string{"health", "farmacia", "medico", "médico", "doctor", "clinica", "clínica", "hospital"}},
	{Name: "Educación", Aliases: []string{"education", "cursos", "curso", "facultad", "universidad", "libros", "libro"}},
	{Name: "Vestimenta", Aliases: []string{"clothing", "ropa", "zapatos", "calzado", "indumentaria"}},
	{Name: "Otros", Aliases: []string{"other", "varios", "general", "misc"}},
}

// CategoryDef defines a single category with its name and known aliases.
type CategoryDef struct {
	Name    string   // Canonical name (Spanish)
	Aliases []string // Acceptable alternatives
}

// NormalizeCategory normalizes a free-text category to its canonical form.
// Returns the canonical name and true if recognized, or the input unchanged if not.
func NormalizeCategory(input string) (string, bool) {
	lower := strings.ToLower(input)
	if lower == "" {
		return "", false
	}
	for _, cat := range PredefinedCategories {
		if strings.ToLower(cat.Name) == lower {
			return cat.Name, true
		}
		for _, alias := range cat.Aliases {
			if strings.ToLower(alias) == lower {
				return cat.Name, true
			}
		}
	}
	return input, false
}

// CategoryList returns a human-readable list of categories for help text.
func CategoryList() string {
	var names []string
	for _, cat := range PredefinedCategories {
		names = append(names, cat.Name)
	}
	if len(names) == 0 {
		return ""
	}
	return strings.Join(names, ", ")
}

// IsDefaultCategory checks if a category is one of the predefined ones.
func IsDefaultCategory(name string) bool {
	_, ok := NormalizeCategory(name)
	return ok
}

// CategorySuggestion returns a suggestion message when an unrecognized category is used.
func CategorySuggestion(input string) string {
	if input == "" {
		return ""
	}
	normalized, ok := NormalizeCategory(input)
	if !ok {
		return normalized // Not recognized, return the original
	}
	if !strings.EqualFold(normalized, input) {
		return normalized // Found via alias — suggest the canonical name
	}
	return "" // Already canonical
}

package role

// Template describes a role with its permissions, e.g. for the role suggestion
// of the Ersteinrichtung.
type Template struct {
	Name        string
	PaysBaseFee bool
	Permissions []TemplatePermission
}

type TemplatePermission struct {
	EntityType     EntityType
	PermissionType PermissionType
}

// clubEntityTypes are the entity types the web UI manages (game_nights and
// penalties are legacy and not shown).
var clubEntityTypes = []EntityType{
	EntityTypeGameDays,
	EntityTypeTransactions,
	EntityTypePlayers,
	EntityTypePenaltyTypes,
	EntityTypeCompetitions,
}

func grant(entities []EntityType, perms ...PermissionType) []TemplatePermission {
	out := make([]TemplatePermission, 0, len(entities)*len(perms))
	for _, e := range entities {
		for _, p := range perms {
			out = append(out, TemplatePermission{EntityType: e, PermissionType: p})
		}
	}
	return out
}

// SuggestedRoles is the role suggestion for a new club: Präsident may do
// everything; Kassenwart manages game days, transactions, players, penalty
// types and competitions and may only read roles; Mitglied and Gast only read
// (Gast pays no base fee). Owners can change everything afterwards.
func SuggestedRoles() []Template {
	all := append([]EntityType{EntityTypeRoles}, clubEntityTypes...)
	read := []PermissionType{PermissionTypeList, PermissionTypeView}
	return []Template{
		{Name: "Präsident", PaysBaseFee: true, Permissions: grant(all, ValidPermissionTypes()...)},
		{Name: "Kassenwart", PaysBaseFee: true, Permissions: append(
			grant(clubEntityTypes, PermissionTypeList, PermissionTypeView, PermissionTypeCreate, PermissionTypeUpdate),
			grant([]EntityType{EntityTypeRoles}, read...)...,
		)},
		{Name: "Mitglied", PaysBaseFee: true, Permissions: grant(clubEntityTypes, read...)},
		{Name: "Gast", PaysBaseFee: false, Permissions: grant(clubEntityTypes, read...)},
	}
}

package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// idFromInt64 formats a numeric SendGrid ID as a Terraform resource ID string.
func idFromInt64(id int64) string {
	return strconv.FormatInt(id, 10)
}

// stringSetToSlice converts a types.Set of strings into a []string. A null
// or unknown set yields a nil slice.
func stringSetToSlice(ctx context.Context, s types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if s.IsNull() || s.IsUnknown() {
		return nil, diags
	}
	var out []string
	diags.Append(s.ElementsAs(ctx, &out, false)...)
	return out, diags
}

// setStringSet builds a types.Set from a []string and stores it in dst.
func setStringSet(ctx context.Context, dst *types.Set, values []string) diag.Diagnostics {
	var diags diag.Diagnostics
	set, d := types.SetValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	*dst = set
	return diags
}

// stringListToSlice converts a types.List of strings into a []string. A null
// or unknown list yields a nil slice.
func stringListToSlice(ctx context.Context, l types.List) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if l.IsNull() || l.IsUnknown() {
		return nil, diags
	}
	var out []string
	diags.Append(l.ElementsAs(ctx, &out, false)...)
	return out, diags
}

// setStringList builds a types.List from a []string and stores it in dst.
func setStringList(ctx context.Context, dst *types.List, values []string) diag.Diagnostics {
	var diags diag.Diagnostics
	list, d := types.ListValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	*dst = list
	return diags
}

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

// idFromInt64 formats a numeric SendGrid ID as a Terraform resource ID string.
func idFromInt64(id int64) string {
	return strconv.FormatInt(id, 10)
}

// clientFromProviderData casts the provider-configured client out of the
// ProviderData handed to every resource/data source Configure method.
func clientFromProviderData(providerData any) (*sg.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	if providerData == nil {
		return nil, diags
	}
	client, ok := providerData.(*sg.Client)
	if !ok {
		diags.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sendgrid.Client, got: %T. This is a bug in the provider.", providerData),
		)
		return nil, diags
	}
	return client, diags
}

// domainAuthenticationDNSMap flattens the DNS records SendGrid returns for a
// domain authentication into the "<record>_type"/"_host"/"_data" map shared
// by the sendgrid_domain_authentication resource and data source.
func domainAuthenticationDNSMap(ctx context.Context, dns sg.DomainAuthenticationDNS) (types.Map, diag.Diagnostics) {
	records := map[string]string{
		"mail_cname_type": dns.MailCNAME.Type,
		"mail_cname_host": dns.MailCNAME.Host,
		"mail_cname_data": dns.MailCNAME.Data,
		"dkim1_type":      dns.DKIM1.Type,
		"dkim1_host":      dns.DKIM1.Host,
		"dkim1_data":      dns.DKIM1.Data,
		"dkim2_type":      dns.DKIM2.Type,
		"dkim2_host":      dns.DKIM2.Host,
		"dkim2_data":      dns.DKIM2.Data,
	}
	return types.MapValueFrom(ctx, types.StringType, records)
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

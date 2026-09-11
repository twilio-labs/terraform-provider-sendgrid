package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

func TestSingleSendParamsFromModelOptionalInt64Pointers(t *testing.T) {
	ctx := context.Background()

	base := func() singleSendResourceModel {
		emptyList, _ := types.ListValueFrom(ctx, types.StringType, []string{})
		return singleSendResourceModel{
			Name:       types.StringValue("test"),
			Categories: emptyList,
			SendTo: &singleSendToModel{
				ListIDs:    emptyList,
				SegmentIDs: emptyList,
				All:        types.BoolValue(false),
			},
			EmailConfig: &singleSendEmailConfigModel{
				Subject:              types.StringValue("hi"),
				GeneratePlainContent: types.BoolValue(true),
				SuppressionGroupID:   types.Int64Null(),
				SenderID:             types.Int64Null(),
			},
		}
	}

	t.Run("both null", func(t *testing.T) {
		m := base()
		p, diags := singleSendParamsFromModel(ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if p.EmailConfig.SuppressionGroupID != nil {
			t.Errorf("SuppressionGroupID = %v, want nil", p.EmailConfig.SuppressionGroupID)
		}
		if p.EmailConfig.SenderID != nil {
			t.Errorf("SenderID = %v, want nil", p.EmailConfig.SenderID)
		}
	})

	t.Run("both set", func(t *testing.T) {
		m := base()
		m.EmailConfig.SuppressionGroupID = types.Int64Value(42)
		m.EmailConfig.SenderID = types.Int64Value(7)
		p, diags := singleSendParamsFromModel(ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if p.EmailConfig.SuppressionGroupID == nil || *p.EmailConfig.SuppressionGroupID != 42 {
			t.Errorf("SuppressionGroupID = %v, want 42", p.EmailConfig.SuppressionGroupID)
		}
		if p.EmailConfig.SenderID == nil || *p.EmailConfig.SenderID != 7 {
			t.Errorf("SenderID = %v, want 7", p.EmailConfig.SenderID)
		}
	})

	t.Run("mixed", func(t *testing.T) {
		m := base()
		m.EmailConfig.SuppressionGroupID = types.Int64Value(42)
		p, diags := singleSendParamsFromModel(ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if p.EmailConfig.SuppressionGroupID == nil || *p.EmailConfig.SuppressionGroupID != 42 {
			t.Errorf("SuppressionGroupID = %v, want 42", p.EmailConfig.SuppressionGroupID)
		}
		if p.EmailConfig.SenderID != nil {
			t.Errorf("SenderID = %v, want nil", p.EmailConfig.SenderID)
		}
	})

	t.Run("nil SendTo and EmailConfig do not panic", func(t *testing.T) {
		m := singleSendResourceModel{Name: types.StringValue("test")}
		if _, diags := singleSendParamsFromModel(ctx, m); diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
	})
}

func TestFlattenSingleSendInt64Pointers(t *testing.T) {
	ctx := context.Background()

	newSingleSend := func(suppressionGroupID, senderID *int64) *sg.SingleSend {
		return &sg.SingleSend{
			ID:     "id1",
			Name:   "test",
			Status: "draft",
			EmailConfig: sg.SingleSendEmailConfig{
				SuppressionGroupID: suppressionGroupID,
				SenderID:           senderID,
			},
		}
	}

	t.Run("nil pointers become Int64Null", func(t *testing.T) {
		var m singleSendResourceModel
		diags := flattenSingleSend(ctx, newSingleSend(nil, nil), &m)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if !m.EmailConfig.SuppressionGroupID.IsNull() {
			t.Errorf("SuppressionGroupID = %v, want null", m.EmailConfig.SuppressionGroupID)
		}
		if !m.EmailConfig.SenderID.IsNull() {
			t.Errorf("SenderID = %v, want null", m.EmailConfig.SenderID)
		}
	})

	t.Run("non-nil pointers become Int64Value", func(t *testing.T) {
		sg1, sn1 := int64(42), int64(7)
		var m singleSendResourceModel
		diags := flattenSingleSend(ctx, newSingleSend(&sg1, &sn1), &m)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if m.EmailConfig.SuppressionGroupID.ValueInt64() != 42 {
			t.Errorf("SuppressionGroupID = %v, want 42", m.EmailConfig.SuppressionGroupID)
		}
		if m.EmailConfig.SenderID.ValueInt64() != 7 {
			t.Errorf("SenderID = %v, want 7", m.EmailConfig.SenderID)
		}
	})

	t.Run("nil SendTo/EmailConfig in model are defensively re-allocated", func(t *testing.T) {
		m := singleSendResourceModel{}
		diags := flattenSingleSend(ctx, newSingleSend(nil, nil), &m)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if m.SendTo == nil || m.EmailConfig == nil {
			t.Fatal("expected SendTo and EmailConfig to be non-nil after flatten")
		}
	})
}

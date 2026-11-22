package model_test

import (
	"errors"
	"testing"

	"helpdesk-backend/internal/model"
)

// TestValidateStatusTransition menguji seluruh matriks aturan transisi status tiket
// sesuai spesifikasi tabel perpindahan status di PDF.
func TestValidateStatusTransition(t *testing.T) {
	tests := []struct {
		name        string
		from        model.TicketStatus
		to          model.TicketStatus
		role        model.UserRole
		wantErr     bool
		expectedErr error
	}{
		// --- Kasus Legal (Harus Berhasil) ---
		{
			name:    "Open -> In Progress oleh Agent (Legal)",
			from:    model.StatusOpen,
			to:      model.StatusInProgress,
			role:    model.RoleAgent,
			wantErr: false,
		},
		{
			name:    "Open -> In Progress oleh Admin (Legal)",
			from:    model.StatusOpen,
			to:      model.StatusInProgress,
			role:    model.RoleAdmin,
			wantErr: false,
		},
		{
			name:    "In Progress -> Resolved oleh Agent (Legal)",
			from:    model.StatusInProgress,
			to:      model.StatusResolved,
			role:    model.RoleAgent,
			wantErr: false,
		},
		{
			name:    "In Progress -> Resolved oleh Admin (Legal)",
			from:    model.StatusInProgress,
			to:      model.StatusResolved,
			role:    model.RoleAdmin,
			wantErr: false,
		},
		{
			name:    "Resolved -> Closed oleh Requester (Legal)",
			from:    model.StatusResolved,
			to:      model.StatusClosed,
			role:    model.RoleRequester,
			wantErr: false,
		},
		{
			name:    "Resolved -> Closed oleh Agent (Legal)",
			from:    model.StatusResolved,
			to:      model.StatusClosed,
			role:    model.RoleAgent,
			wantErr: false,
		},
		{
			name:    "Resolved -> Closed oleh Admin (Legal)",
			from:    model.StatusResolved,
			to:      model.StatusClosed,
			role:    model.RoleAdmin,
			wantErr: false,
		},
		{
			name:    "Resolved -> Open (Reopen) oleh Requester (Legal)",
			from:    model.StatusResolved,
			to:      model.StatusOpen,
			role:    model.RoleRequester,
			wantErr: false,
		},
		{
			name:    "Open -> Closed (Batal) oleh Requester (Legal)",
			from:    model.StatusOpen,
			to:      model.StatusClosed,
			role:    model.RoleRequester,
			wantErr: false,
		},
		{
			name:    "Open -> Closed oleh Admin (Legal)",
			from:    model.StatusOpen,
			to:      model.StatusClosed,
			role:    model.RoleAdmin,
			wantErr: false,
		},

		// --- Kasus Ilegal (Harus Ditolak) ---
		{
			name:        "Open -> In Progress oleh Requester (Ilegal: requester tidak boleh mengerjakan tiket)",
			from:        model.StatusOpen,
			to:          model.StatusInProgress,
			role:        model.RoleRequester,
			wantErr:     true,
			expectedErr: model.ErrIllegalTransition,
		},
		{
			name:        "Open -> Closed oleh Agent (Ilegal: hanya requester & admin yang boleh membatalkan)",
			from:        model.StatusOpen,
			to:          model.StatusClosed,
			role:        model.RoleAgent,
			wantErr:     true,
			expectedErr: model.ErrIllegalTransition,
		},
		{
			name:        "In Progress -> Closed langsung (Ilegal: harus melalui status Resolved)",
			from:        model.StatusInProgress,
			to:          model.StatusClosed,
			role:        model.RoleAgent,
			wantErr:     true,
			expectedErr: model.ErrIllegalTransition,
		},
		{
			name:        "Resolved -> Open oleh Agent (Ilegal: hanya requester yang boleh membuka kembali)",
			from:        model.StatusResolved,
			to:          model.StatusOpen,
			role:        model.RoleAgent,
			wantErr:     true,
			expectedErr: model.ErrIllegalTransition,
		},
		{
			name:        "Resolved -> Open oleh Admin (Ilegal: hanya requester yang boleh membuka kembali)",
			from:        model.StatusResolved,
			to:          model.StatusOpen,
			role:        model.RoleAdmin,
			wantErr:     true,
			expectedErr: model.ErrIllegalTransition,
		},
		{
			name:        "Closed -> Open (Ilegal: tiket Closed bersifat terminal dan tidak bisa dibuka lagi)",
			from:        model.StatusClosed,
			to:          model.StatusOpen,
			role:        model.RoleRequester,
			wantErr:     true,
			expectedErr: model.ErrIllegalTransition,
		},
		{
			name:        "Status yang sama: Open -> Open (Ilegal)",
			from:        model.StatusOpen,
			to:          model.StatusOpen,
			role:        model.RoleAdmin,
			wantErr:     true,
			expectedErr: model.ErrSameStatus,
		},
		{
			name:        "Status tujuan tidak valid (Ilegal)",
			from:        model.StatusOpen,
			to:          model.TicketStatus("UnknownStatus"),
			role:        model.RoleAdmin,
			wantErr:     true,
			expectedErr: model.ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := model.ValidateStatusTransition(tt.from, tt.to, tt.role)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateStatusTransition() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("ValidateStatusTransition() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}
		})
	}
}

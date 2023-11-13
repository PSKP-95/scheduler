package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	mockdb "github.com/PSKP-95/scheduler/db/mock"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_getScheduleHistory(t *testing.T) {
	sched_id := uuid.New()
	history := createHistory(sched_id)

	testCases := []struct {
		name       string
		id         string
		page       string
		size       string
		buildStubs func(store *mockdb.MockStore)
		respCode   int
	}{
		{
			name: "correct id",
			id:   sched_id.String(),
			page: "1",
			size: "10",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListHistory(gomock.Any(), gomock.Any()).Times(1).Return(history, nil)
			},
			respCode: http.StatusOK,
		},
		{
			name:       "empty id",
			id:         "",
			page:       "1",
			size:       "10",
			buildStubs: func(store *mockdb.MockStore) {},
			respCode:   http.StatusNotFound,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)

			url := fmt.Sprintf("/api/schedule/%s/history", tc.id)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			resp, _ := server.app.Test(request, 5)
			require.Equal(t, resp.StatusCode, tc.respCode)
		})
	}
}

func createHistory(id uuid.UUID) []db.ListHistoryRow {
	history := []db.ListHistoryRow{
		{
			OccurenceID:  1,
			Schedule:     id,
			Status:       db.StatusFailure,
			Manual:       false,
			ScheduledAt:  time.Now(),
			StartedAt:    time.Now().Add(1 * time.Second),
			CompletedAt:  time.Now().Add(2 * time.Second),
			TotalRecords: 2,
		},
		{
			OccurenceID:  2,
			Schedule:     id,
			Status:       db.StatusSuccess,
			Manual:       false,
			ScheduledAt:  time.Now(),
			StartedAt:    time.Now().Add(1 * time.Second),
			CompletedAt:  time.Now().Add(2 * time.Second),
			TotalRecords: 2,
		},
	}
	return history
}

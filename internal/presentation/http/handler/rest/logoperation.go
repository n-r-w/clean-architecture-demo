package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/n-r-w/log-server-v2/internal/domain/entity"
	"github.com/n-r-w/log-server-v2/internal/presentation/http/handler"
	schema_log "github.com/n-r-w/log-server-v2/internal/schema/schema.log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Добавить в лог
func (info *restInfo) addLogRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []entity.LogRecord
		// парсим входящий json
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			info.controller.RespondError(w, http.StatusBadRequest, err)

			return
		}

		if err := info.log.Insert(req); err != nil {
			info.controller.RespondError(w, http.StatusForbidden, err)

			return
		}

		info.controller.RespondData(w, http.StatusCreated, nil)
	}
}

// Получить записи из лога
func (info *restInfo) getLogRecords() http.HandlerFunc {
	type requestParams struct {
		TimeFrom time.Time `json:"timeFrom"`
		TimeTo   time.Time `json:"timeTo"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &requestParams{
			TimeFrom: time.Time{},
			TimeTo:   time.Time{},
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			info.controller.RespondError(w, http.StatusBadRequest, err)

			return
		}

		records, _, err := info.log.Find(req.TimeFrom, req.TimeTo, info.maxLogRecordsResult)
		if err != nil {
			info.controller.RespondError(w, http.StatusInternalServerError, err)

			return
		}

		if len(records) == 0 {
			info.controller.RespondData(w, http.StatusOK, nil)

			return
		}

		if r.Header.Get(binaryFormatHeaderName) == binaryFormatHeaderProtobuf {
			// клиент хочет Protobuf
			mRecords := &schema_log.LogRecords{
				Records: nil,
			}

			for _, r := range records {
				mRecord := &schema_log.LogRecord{
					Id:       r.ID,
					LogTime:  timestamppb.New(r.LogTime),
					RealTime: timestamppb.New(r.RealTime),
					Level:    uint32(r.Level),
					Message1: r.Message1,
					Message2: r.Message2,
					Message3: r.Message3,
				}
				mRecords.Records = append(mRecords.Records, mRecord)
			}

			out, err := proto.Marshal(mRecords)
			if err != nil {
				info.controller.RespondError(w, http.StatusInternalServerError, err)

				return
			}

			w.Header().Add(binaryFormatHeaderName, binaryFormatHeaderProtobuf)
			info.controller.RespondCompressed(w, r, http.StatusOK, handler.CompressionGzip, out)

			return
		}

		// отдаем с gzip сжатием если клиент это желает
		info.controller.RespondCompressed(w, r, http.StatusOK, handler.CompressionGzip, &records)
	}
}

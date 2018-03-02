package webapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"io/ioutil"

	"mime/multipart"

	"bitbucket.org/digitorus/pdfsigner/queued_sign"
	"bitbucket.org/digitorus/pdfsigner/signer"
	"github.com/gorilla/mux"
	errors2 "github.com/pkg/errors"
)

func (wa *WebAPI) handleSignSchedule(w http.ResponseWriter, r *http.Request) {
	// put job with specified signer
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var f fields
	var fileNames []string

	for {
		// get part
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			httpError(w, errors2.Wrap(err, "get multipart"), 500)
			return
		}

		//parse fields
		err = parseFields(p, &f)
		if err != nil {
			httpError(w, errors2.Wrap(err, "parse fields"), 500)
			return
		}

		//save pdf file to tmp
		err = savePDFToTemp(p, &fileNames)
		if err != nil {
			httpError(w, errors2.Wrap(err, "save pdf to tmp"), 500)
			return
		}
	}

	sessionID, err := pushSignJob(wa.qSign, f, fileNames)
	if err != nil {
		httpError(w, errors2.Wrap(err, "push jobs"), 500)
		return
	}

	_, err = fmt.Fprint(w, sessionID)
	if err != nil {
		log.Println(err)
	}
}

func (wa *WebAPI) handleSignCheck(w http.ResponseWriter, r *http.Request) {
	// get jobs for session
	vars := mux.Vars(r)
	sessionId := vars["sessionID"]

	sess, err := wa.qSign.GetSessionByID(sessionId)
	if err != nil {
		httpError(w, err, 500)
		return
	}

	// respond with json
	j, err := json.Marshal(sess)
	if err != nil {
		httpError(w, err, 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)

}

func (wa *WebAPI) handleSignGetFile(w http.ResponseWriter, r *http.Request) {
	// get jobs for session
	vars := mux.Vars(r)
	sessionId := vars["sessionID"]
	fileID := vars["fileID"]

	// get file path
	filePath, err := wa.qSign.GetCompletedJobFilePath(sessionId, fileID)
	if err != nil {
		httpError(w, err, 500)
		return
	}

	// get file
	file, err := os.Open(filePath)
	if err != nil {
		httpError(w, err, 500)
		return
	}
	defer file.Close()

	// get file info
	fileInfo, err := file.Stat()
	if err != nil {
		httpError(w, err, 500)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	_, err = io.Copy(w, file)
	if err != nil {
		httpError(w, err, 500)
		return
	}
}

type fields struct {
	signerName string
	signData   signer.SignData
}

func parseFields(p *multipart.Part, f *fields) error {
	switch p.FormName() {
	case "signer", "name", "location", "reason", "contactInfo", "certType", "approval":
		//parse params
		slurp, err := ioutil.ReadAll(p)
		if err != nil {
			return nil
		}
		str := string(slurp)

		switch p.FormName() {
		case "signer":
			f.signerName = str
		case "name":
			f.signData.Signature.Info.Name = str
		case "location":
			f.signData.Signature.Info.Location = str
		case "reason":
			f.signData.Signature.Info.Reason = str
		case "contactInfo":
			f.signData.Signature.Info.ContactInfo = str
		case "certType":
			i, err := strconv.Atoi(str)
			if err != nil {
				return err
			}
			f.signData.Signature.CertType = uint32(i)
		case "approval":
			b, err := strconv.ParseBool(str)
			if err != nil {
				return err
			}
			f.signData.Signature.Approval = b
		}
	}

	return nil
}

func (wa *WebAPI) handleSignDelete(w http.ResponseWriter, r *http.Request) {
	// get jobs for session
	vars := mux.Vars(r)
	sessionId := vars["sessionID"]
	wa.qSign.DeleteSession(sessionId)
}

func pushSignJob(qs *queued_sign.QSign, f fields, fileNames []string) (string, error) {
	if f.signerName == "" {
		return "", errors.New("signer name is required")
	}

	totalJobs := len(fileNames)

	sessionID := qs.NewSession(totalJobs, f.signData)
	priority := determinePriority(totalJobs)

	for _, fileName := range fileNames {
		_, err := qs.PushJob(f.signerName, sessionID, fileName, fileName+"_signed.pdf", priority)
		if err != nil {
			return "", err
		}
	}

	return sessionID, nil
}

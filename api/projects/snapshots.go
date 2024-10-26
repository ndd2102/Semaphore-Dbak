package projects

import (
    "encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
    "strings"

	"github.com/ansible-semaphore/semaphore/api/helpers"
    "github.com/ansible-semaphore/semaphore/db"
)

// DecryptPassword giải mã và trả về mật khẩu từ AccessKey đã giải mã
func DecryptPassword(key *db.AccessKey) (string, error) {
	// Giải mã Secret từ AccessKey
	err := key.DeserializeSecret()
	if err != nil {
		return "", fmt.Errorf("Failed to decrypt secret: %v", err)
	}

	// Kiểm tra loại AccessKey để lấy mật khẩu
	if key.Type == db.AccessKeyLoginPassword {
		return key.LoginPassword.Password, nil
	}

	return "", fmt.Errorf("AccessKey type is not 'login_password'")
}

// GetSnapshotData lấy dữ liệu từ repository của Restic
func GetSnapshotData(w http.ResponseWriter, r *http.Request) {
	// Lấy tham số từ query string
	projectIDStr := r.URL.Query().Get("project_id")
	accessKeyIDStr := r.URL.Query().Get("access_key_id")

	// Kiểm tra xem các tham số có hợp lệ hay không
	if projectIDStr == "" || accessKeyIDStr == "" {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Missing project_id or access_key_id",
		})
		return
	}

	// Chuyển đổi projectID và accessKeyID từ chuỗi sang số nguyên
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid project_id",
		})
		return
	}

	accessKeyID, err := strconv.Atoi(accessKeyIDStr)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid access_key_id",
		})
		return
	}

	// Debug: In giá trị ra để kiểm tra
	fmt.Printf("projectID: %d, accessKeyID: %d\n", projectID, accessKeyID)

	// Lấy thông tin AccessKey từ cơ sở dữ liệu
	key, err := helpers.Store(r).GetAccessKey(projectID, accessKeyID) // Gọi hàm GetAccessKey từ Store
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to get access key: " + err.Error(),
		})
		return
	}

	// Giải mã mật khẩu từ secret
	password, err := DecryptPassword(&key)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to decrypt password: " + err.Error(),
		})
		return
	}

	// Debug: In giá trị password đã giải mã
	fmt.Printf("Decrypted password: %s\n", password)

	// Thiết lập các biến môi trường cần thiết cho Restic
	os.Setenv("RESTIC_REPOSITORY", "s3:https://xplat-minio-api.dev.cluster02.fis-cloud.xplat.online/backup")
	os.Setenv("RESTIC_PASSWORD", "asdf")        // Sử dụng mật khẩu đã giải mã từ AccessKey
	os.Setenv("AWS_ACCESS_KEY_ID", key.Name)      // Name của AccessKey là AWS_ACCESS_KEY_ID
	os.Setenv("AWS_SECRET_ACCESS_KEY", password) // Secret đã giải mã là AWS_SECRET_ACCESS_KEY

	// Ví dụ lệnh để lấy snapshots từ repository của Restic
	cmd := exec.Command("restic", "snapshots", "--json")

	// Thực thi lệnh và lấy kết quả
	output, err := cmd.CombinedOutput()
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to execute Restic command: " + err.Error(),
		})
		return
	}

	// Trả lại dữ liệu lệnh đã thực thi dưới dạng JSON
	var snapshots []map[string]interface{}
	err = json.Unmarshal(output, &snapshots)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to parse JSON output: " + err.Error(),
		})
		return
	}

	for _, snapshot := range snapshots {
		snapshotID, ok := snapshot["id"].(string)
		if !ok {
			continue
		}

		// Chạy lệnh để lấy kích thước snapshot
		statsCmd := exec.Command("restic", "stats", snapshotID, "--mode", "restore-size")
		statsOutput, err := statsCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to get stats for snapshot %s: %s\n", snapshotID, err.Error())
			continue
		}
        fmt.Printf("Stats output for snapshot %s:\n%s\n", snapshotID, string(statsOutput))
		// Lấy kích thước từ đầu ra của lệnh
        // statsStr := strings.TrimSpace(string(statsOutput))
		for _, line := range strings.Split(string(statsOutput), "\n") {
			if strings.Contains(line, "Total Size:") {
				// Tách lấy phần kích thước bao gồm cả đơn vị MiB, GiB
				size := strings.TrimSpace(strings.Split(line, "Total Size:")[1])
				snapshot["size"] = size
				break
            }
		}
	}

	helpers.WriteJSON(w, http.StatusOK, snapshots)
}

// DeleteSnapshot xóa một snapshot từ repository của Restic
func DeleteSnapshot(w http.ResponseWriter, r *http.Request) {
	// Lấy tham số từ query string
	projectIDStr := r.URL.Query().Get("project_id")
	accessKeyIDStr := r.URL.Query().Get("access_key_id")
	snapshotID := r.URL.Query().Get("snapshot_id")

	// Kiểm tra xem các tham số có hợp lệ không
	if projectIDStr == "" || accessKeyIDStr == "" || snapshotID == "" {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Missing project_id, access_key_id, or snapshot_id",
		})
		return
	}

	// Chuyển đổi projectID và accessKeyID từ chuỗi sang số nguyên
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid project_id",
		})
		return
	}

	accessKeyID, err := strconv.Atoi(accessKeyIDStr)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid access_key_id",
		})
		return
	}

	// Debug: In giá trị ra để kiểm tra
	fmt.Printf("projectID: %d, accessKeyID: %d, snapshotID: %s\n", projectID, accessKeyID, snapshotID)

	// Lấy thông tin AccessKey từ cơ sở dữ liệu
	key, err := helpers.Store(r).GetAccessKey(projectID, accessKeyID)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to get access key: " + err.Error(),
		})
		return
	}

	// Giải mã secret để lấy mật khẩu
	err = key.DeserializeSecret()
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to decrypt secret: " + err.Error(),
		})
		return
	}

	// Thiết lập các biến môi trường cần thiết cho Restic
	os.Setenv("RESTIC_REPOSITORY", "s3:https://restic.com")
	os.Setenv("RESTIC_PASSWORD", *key.Secret)
	os.Setenv("AWS_ACCESS_KEY_ID", key.Name)
	os.Setenv("AWS_SECRET_ACCESS_KEY", *key.Secret)

	// Lệnh xóa snapshot bằng restic
	cmd := exec.Command("restic", "forget", snapshotID)

	// Thực thi lệnh và kiểm tra kết quả
	output, err := cmd.CombinedOutput()
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete snapshot: " + err.Error(),
			"output": string(output),
		})
		return
	}

	// Trả lại thông báo thành công
	helpers.WriteJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Snapshot %s deleted successfully.", snapshotID),
	})
}
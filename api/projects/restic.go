package projects

import (
    "net/http"
    "errors"
    "os"
    "os/exec"
    "strings"
    "fmt"
    "encoding/json"

    "github.com/ansible-semaphore/semaphore/api/helpers"
    "github.com/ansible-semaphore/semaphore/db"
    "github.com/gorilla/context"
)

// ResticConfigMiddleware ensures that the Restic configuration exists and loads it into the context
func ResticConfigMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        project := context.Get(r, "project").(db.Project)
        restic_configID, err := helpers.GetIntParam("restic_config_id", w, r)
        if err != nil {
            helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
                "error": "Invalid or missing 'restic_config_id' parameter",
            })
            return
        }

        restic_config, err := helpers.Store(r).GetResticConfig(project.ID, restic_configID)
        
        if err != nil {
            helpers.WriteError(w, err)
            return
        }

        context.Set(r, "restic_config", restic_config)
        next.ServeHTTP(w, r)
    })
}

// GetResticConfigs returns a list of Restic configurations in the project
func GetResticConfigs(w http.ResponseWriter, r *http.Request) {
    if restic := context.Get(r, "restic_config"); restic != nil {
        helpers.WriteJSON(w, http.StatusOK, restic.(db.ResticConfig))
        return
    }

    project := context.Get(r, "project").(db.Project)
	restics, err := helpers.Store(r).GetResticConfigs(project.ID, helpers.QueryParams(r.URL))

    if err != nil {
		helpers.WriteError(w, err)
		return
	}

    helpers.WriteJSON(w, http.StatusOK, restics)
}

// AddResticConfig creates a new Restic configuration
func AddResticConfig(w http.ResponseWriter, r *http.Request) {
    project := context.Get(r, "project").(db.Project)

    var restic_config db.ResticConfig

    if !helpers.Bind(w, r, &restic_config) {
        return
    }

    if restic_config.ProjectID != project.ID {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
            "error": "Project ID in body and URL must be the same",
        })
    }

    newRestic, err := helpers.Store(r).CreateResticConfig(restic_config)
    
    if err != nil {
        helpers.WriteError(w, err)
        return
    }

    helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
        UserID:      helpers.UserFromContext(r).ID,
        ProjectID:   newRestic.ProjectID,
        ObjectType:  db.EventResticConfig,
        ObjectID:    newRestic.ID,
        Description: "Restic configuration created",
    })

    w.WriteHeader(http.StatusNoContent)
}

// UpdateResticConfig updates an existing Restic configuration
func UpdateResticConfig(w http.ResponseWriter, r *http.Request) {
    oldRestic := context.Get(r, "restic_config").(db.ResticConfig)
    var restic_config db.ResticConfig

    if !helpers.Bind(w, r, &restic_config) {
        return
    }

    if restic_config.ID != oldRestic.ID {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
            "error": "Restic ID in body and URL must be the same",
        })
        return
    }

    if restic_config.ProjectID != oldRestic.ProjectID {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
            "error": "Project ID in body and URL must be the same",
        })
        return
    }

    if err := helpers.Store(r).UpdateResticConfig(restic_config); err != nil {
        helpers.WriteError(w, err)
        return
    }

    helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
        UserID:      helpers.UserFromContext(r).ID,
        ProjectID:   oldRestic.ProjectID,
        ObjectType:  db.EventResticConfig,
        ObjectID:    oldRestic.ID,
        Description: "Restic configuration updated",
    })

    w.WriteHeader(http.StatusNoContent)
}

// RemoveResticConfig deletes a Restic configuration
func RemoveResticConfig(w http.ResponseWriter, r *http.Request) {
    restic_config := context.Get(r, "restic_config").(db.ResticConfig)

    var err error

    err = helpers.Store(r).DeleteResticConfig(restic_config.ProjectID, restic_config.ID)
	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Restic is in use by one or more templates",
			"inUse": true,
		})
		return
	}

    if err != nil {
		helpers.WriteError(w, err)
		return
	}
    
    helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
        UserID:      helpers.UserFromContext(r).ID,
        ProjectID:   restic_config.ProjectID,
        ObjectType:  db.EventResticConfig,
        ObjectID:    restic_config.ID,
        Description: "Restic configuration deleted",
    })

    w.WriteHeader(http.StatusNoContent)
}

func GetSnapshotData(w http.ResponseWriter, r *http.Request) {
    project := context.Get(r, "project").(db.Project)
    restic_config := context.Get(r, "restic_config").(db.ResticConfig)
    restic_configID, err := helpers.GetIntParam("restic_config_id", w, r)
    if err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
            "error": "Invalid or missing 'restic_config_id' parameter",
        })
        return
    }

    key, err := helpers.Store(r).GetAccessKey(project.ID, restic_configID) // Gọi hàm GetAccessKey từ Store
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to get access key: " + err.Error(),
		})
		return
	}

    err = key.DeserializeSecret()
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to decrypt secret: " + err.Error(),
		})
		return
	}

    var password string
    var username string
	if key.Type == db.AccessKeyLoginPassword {
        username = key.LoginPassword.Login
		password = key.LoginPassword.Password
	} else {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "AccessKey type is not 'login_password'",
		})
		return
	}

    // Set environment variables needed for Restic using retrieved config data
    os.Setenv("RESTIC_REPOSITORY", "s3:" + restic_config.URL + "/" + restic_config.Bucket)
    os.Setenv("RESTIC_PASSWORD", restic_config.ResticKey)        
    os.Setenv("AWS_ACCESS_KEY_ID", username)    
    os.Setenv("AWS_SECRET_ACCESS_KEY", password)

	// Debug: In giá trị ra để kiểm tra
	fmt.Printf("projectID: %s, accessKeyID: %s, keyName: %s, password: %s\n", restic_config.URL + "/" + restic_config.Bucket, restic_config.ResticKey, username, password)

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

		for _, line := range strings.Split(string(statsOutput), "\n") {
			if strings.Contains(line, "Total Size:") {
				size := strings.TrimSpace(strings.Split(line, "Total Size:")[1])
				snapshot["size"] = size
				break
            }
		}
	}

	helpers.WriteJSON(w, http.StatusOK, snapshots)
}

func RemoveSnapshot(w http.ResponseWriter, r *http.Request) {
    project := context.Get(r, "project").(db.Project)
    restic_config := context.Get(r, "restic_config").(db.ResticConfig)
    restic_configID, err := helpers.GetIntParam("restic_config_id", w, r)
    if err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
            "error": "Invalid or missing 'restic_config_id' parameter",
        })
        return
    }

    // Lấy snapshotID từ URL hoặc request body
    snapshotID := r.URL.Query().Get("snapshot_id")
    if snapshotID == "" {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
            "error": "Missing 'snapshot_id' parameter",
        })
        return
    }

    key, err := helpers.Store(r).GetAccessKey(project.ID, restic_configID)
    if err != nil {
        helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
            "error": "Failed to get access key: " + err.Error(),
        })
        return
    }

    // Giải mã Secret từ AccessKey trực tiếp
    err = key.DeserializeSecret()
    if err != nil {
        helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
            "error": "Failed to decrypt secret: " + err.Error(),
        })
        return
    }

    // Kiểm tra loại AccessKey để lấy mật khẩu
    var password string
    if key.Type == db.AccessKeyLoginPassword {
        password = key.LoginPassword.Password
    } else {
        helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
            "error": "AccessKey type is not 'login_password'",
        })
        return
    }

    // Set environment variables needed for Restic
    os.Setenv("RESTIC_REPOSITORY", "s3:" + restic_config.URL + "/" + restic_config.Bucket)
    os.Setenv("RESTIC_PASSWORD", restic_config.ResticKey)
    os.Setenv("AWS_ACCESS_KEY_ID", key.Name)
    os.Setenv("AWS_SECRET_ACCESS_KEY", password)

    // Xóa snapshot bằng lệnh Restic
    cmd := exec.Command("restic", "forget", snapshotID, "--prune")
    output, err := cmd.CombinedOutput()
    if err != nil {
        helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
            "error": "Failed to delete snapshot: " + err.Error(),
            "details": string(output),
        })
        return
    }

    // Phản hồi khi xóa thành công
    helpers.WriteJSON(w, http.StatusOK, map[string]string{
        "message": "Snapshot deleted successfully",
        "snapshot_id": snapshotID,
    })
}

func GetResticCredentials(w http.ResponseWriter, r *http.Request) {
    project := context.Get(r, "project").(db.Project)
    restic_configID, err := helpers.GetIntParam("restic_config_id", w, r)
    if err != nil {
        helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
            "error": "Invalid or missing 'restic_config_id' parameter",
        })
        return
    }

    key, err := helpers.Store(r).GetAccessKey(project.ID, restic_configID)
    if err != nil {
        helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
            "error": "Failed to get access key: " + err.Error(),
        })
        return
    }

    // Giải mã Secret từ AccessKey trực tiếp
    err = key.DeserializeSecret()
    if err != nil {
        helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{
            "error": "Failed to decrypt secret: " + err.Error(),
        })
        return
    }

    // Trả về username và password
    helpers.WriteJSON(w, http.StatusOK, map[string]string{
        "username": key.LoginPassword.Login,
        "password": key.LoginPassword.Password,
    })
}
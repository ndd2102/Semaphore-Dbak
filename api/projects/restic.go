package projects

import (
    "net/http"
    "errors"

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

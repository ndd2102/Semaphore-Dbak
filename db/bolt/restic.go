package bolt

import (
    "github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) GetResticConfig(projectID int, restic_configID int) (restic_config db.ResticConfig, err error) {
    err = d.getObject(projectID, db.ResticConfigProps, intObjectID(restic_configID), &restic_config)
	if err != nil {
		return
	}
	restic_config.SSHKey, err = d.GetAccessKey(projectID, restic_config.SSHKeyID)
    return
}

func (d *BoltDb) GetResticConfigs(projectID int, params db.RetrieveQueryParams) (restic_configs []db.ResticConfig, err error) {
    err = d.getObjects(projectID, db.ResticConfigProps, params, nil, &restic_configs)
    return
}

func (d *BoltDb) CreateResticConfig(restic_config db.ResticConfig) (db.ResticConfig, error) {
    err := restic_config.Validate()
    if err != nil {
        return db.ResticConfig{}, err
    }

    newRestic, err := d.createObject(restic_config.ProjectID, db.ResticConfigProps, restic_config)
    return newRestic.(db.ResticConfig), err
}

func (d *BoltDb) UpdateResticConfig(restic_config db.ResticConfig) error {
    err := restic_config.Validate()
    if err != nil {
        return err
    }

    return d.updateObject(restic_config.ProjectID, db.ResticConfigProps, restic_config)
}

func (d *BoltDb) DeleteResticConfig(projectID int, restic_configID int) error {
    return d.deleteObject(projectID, db.ResticConfigProps, intObjectID(restic_configID), nil)
}

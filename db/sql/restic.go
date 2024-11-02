package sql

import (
    "github.com/Masterminds/squirrel"
    "github.com/ansible-semaphore/semaphore/db"
)

func (d *SqlDb) GetResticConfig(projectID int, configID int) (db.ResticConfig, error) {
    var restic_config db.ResticConfig
    err := d.getObject(projectID, db.ResticConfigProps, configID, &restic_config)
    
    if err != nil {
		return restic_config, err
	}

	restic_config.SSHKey, err = d.GetAccessKey(projectID, restic_config.SSHKeyID)

	return restic_config, err
}

func (d *SqlDb) GetResticConfigs(projectID int, params db.RetrieveQueryParams) (restic_configs []db.ResticConfig, err error) {
	q := squirrel.Select("*").
		From("project__restic_config pr")

	order := "ASC"
	if params.SortInverted {
		order = "DESC"
	}

	switch params.SortBy {
	case "name", "git_url":
		q = q.Where("pr.project_id=?", projectID).
			OrderBy("pr." + params.SortBy + " " + order)
	case "ssh_key":
		q = q.LeftJoin("access_key ak ON (pr.ssh_key_id = ak.id)").
			Where("pr.project_id=?", projectID).
			OrderBy("ak.name " + order)
	default:
		q = q.Where("pr.project_id=?", projectID).
			OrderBy("pr.name " + order)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&restic_configs, query, args...)

	return
}

func (d *SqlDb) CreateResticConfig(restic_config db.ResticConfig) (newRestic db.ResticConfig, err error) {
    err = restic_config.Validate()
    
    if err != nil {
        return db.ResticConfig{}, err
    }

    insertID, err := d.insert(
        "id",
        "INSERT INTO project__restic_config (project_id, name, url, restic_key, ssh_key_id, bucket) VALUES (?, ?, ?, ?, ?, ?)",
        restic_config.ProjectID,
        restic_config.Name,
        restic_config.URL,
        restic_config.ResticKey,
        restic_config.SSHKeyID,
        restic_config.Bucket)

    if err != nil {
        return db.ResticConfig{}, err
    }

    newRestic = restic_config
    restic_config.ID = insertID
    return
}

func (d *SqlDb) UpdateResticConfig(restic_config db.ResticConfig) error {
    err := restic_config.Validate()
    if err != nil {
        return err
    }

    _, err = d.exec(
        "UPDATE project__restic_config SET name=?, url=?, restic_key=?, ssh_key_id=?, bucket=? WHERE id=? AND project_id=?",
        restic_config.Name,
        restic_config.URL,
        restic_config.ResticKey,
        restic_config.SSHKeyID,
        restic_config.Bucket,
        restic_config.ID,
        restic_config.ProjectID)

    return err
}

func (d *SqlDb) DeleteResticConfig(projectID int, restic_configID int) error {
    return d.deleteObject(projectID, db.ResticConfigProps, restic_configID)
}
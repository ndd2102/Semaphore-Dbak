package sql

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *SqlDb) GetEnvironment(projectID int, environmentID int) (environment db.Environment, err error) {
	err = d.getObject(projectID, db.EnvironmentProps, environmentID, &environment)
	return
}

func (d *SqlDb) GetEnvironmentRefs(projectID int, environmentID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.EnvironmentProps, environmentID)
}

func (d *SqlDb) GetEnvironments(projectID int, params db.RetrieveQueryParams) ([]db.Environment, error) {
	var environment []db.Environment
	err := d.getProjectObjects(projectID, db.EnvironmentProps, params, &environment)
	return environment, err
}

func (d *SqlDb) UpdateEnvironment(env db.Environment) error {
	err := env.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__environment set name=?, json=?, env=?, password=? where id=?",
		env.Name,
		env.JSON,
		env.ENV,
		env.Password,
		env.ID)
	return err
}

func (d *SqlDb) CreateEnvironment(env db.Environment) (newEnv db.Environment, err error) {
	err = env.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__environment (project_id, name, json, env, password) values (?, ?, ?, ?, ?)",
		env.ProjectID,
		env.Name,
		env.JSON,
		env.ENV,
		env.Password)

	if err != nil {
		return
	}

	newEnv = env
	newEnv.ID = insertID
	return
}

func (d *SqlDb) DeleteEnvironment(projectID int, environmentID int) error {
	return d.deleteObject(projectID, db.EnvironmentProps, environmentID)
}

package model

import (
	"encoding/json"
	"gateway/docker"
	"gateway/logreport"
	apsql "gateway/sql"
)

type dockerRemoteEndpointNotificationListener struct {
	*apsql.DB
}

// StartDockerEndpointUpdateListener registers a listener for updates from Docker remote endpoints.
func StartDockerEndpointUpdateListener(db *apsql.DB) {
	listener := &dockerRemoteEndpointNotificationListener{db}
	db.RegisterListener(listener)
}

// Notify tells the listener a particular notification was fired
func (l *dockerRemoteEndpointNotificationListener) Notify(n *apsql.Notification) {
	if n.Table != "remote_endpoints" {
		return
	}

	switch n.Event {
	case apsql.Update, apsql.Insert:
		remoteEndpoint, err := FindRemoteEndpointForAPIIDAndAccountID(l.DB,
			n.ID, n.APIID, n.AccountID)

		if err != nil {
			logreport.Printf("%s Error finding RemoteEndpointID: %d for APIID: %d and AccountID: %d", err, n.ID, n.APIID, n.AccountID)
			return
		}

		if remoteEndpoint.Type != RemoteEndpointTypeDocker {
			return
		}
		l.handleNotification(n, remoteEndpoint)
	}
}

// Reconnect tells the listener that we may have been disconnected, but
// have reconnected. They should update all state that could have changed.
func (l *dockerRemoteEndpointNotificationListener) Reconnect() {
	// Nothing to do here
}

func (l *dockerRemoteEndpointNotificationListener) handleNotification(n *apsql.Notification, dre *RemoteEndpoint) {
	dc := new(docker.DockerConfig)
	if err := json.Unmarshal(dre.Data, dc); err != nil {
		logreport.Printf("Unable to unmarshal endpoint data: %v", err)
		return
	}
	if err := l.UpdateLogic(dc, dre); err != nil {
		logreport.Printf("%s Error handling docker remote endpoint notification", err)
		return
	}
	for _, v := range dre.EnvironmentData {
		dec := new(docker.DockerConfig)
		if err := json.Unmarshal(v.Data, dec); err != nil {
			logreport.Printf("Unable to unmarshal endpoint environment data: %v", err)
			return
		}
		if err := l.UpdateLogic(dec, dre); err != nil {
			logreport.Printf("%s Error handling docker remote endpoint notification", err)
			return
		}
	}
	return
}

func (l *dockerRemoteEndpointNotificationListener) UpdateLogic(dc *docker.DockerConfig, re *RemoteEndpoint) error {
	UpdateDockerStatus(l.DB, re, RemoteEndpointStatusProcessing, "")
	err := dc.PullOrRefresh()
	if err != nil {
		logreport.Printf("%s Error pulling docker image %s for remote endpoint %d", err, dc.Image(), re.ID)
		return UpdateDockerStatus(l.DB, re, RemoteEndpointStatusFailed, err.Error())
	}
	return UpdateDockerStatus(l.DB, re, RemoteEndpointStatusSuccess, "Pulled image.")
}

func UpdateDockerStatus(db *apsql.DB, re *RemoteEndpoint, status string, statusMessage string) error {
	re.Status = apsql.MakeNullString(status)
	if statusMessage == "" {
		re.StatusMessage = apsql.MakeNullStringNull()
	} else {
		re.StatusMessage = apsql.MakeNullString(statusMessage)
	}
	// In a new transaction, update the status to processing before we do anything
	err := db.DoInTransaction(func(tx *apsql.Tx) error {
		return re.UpdateStatus(tx)
	})

	if err != nil {
		logreport.Printf("Unable to update status to %v.", status)
	}

	return err
}

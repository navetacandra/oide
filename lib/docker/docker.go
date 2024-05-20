package docker

import (
  "fmt"
	"database/sql"
	"os/exec"
)

func getOutput(command string, args ...string) (string, error) {
  cmd := exec.Command(command, args...)
  res, err := cmd.Output()
  return string(res), err
}

func CreatePlaygroundContainer(db *sql.DB, username string) bool {
  var ip string
  var port int
  
  tx, err := db.Begin()
  if err != nil {
    fmt.Printf("Error start transaction: %v\n", err)
    return false
  }

  row := tx.QueryRow("SELECT MAX(container_ip)::inet+1 AS ip, MAX(ssh_port)+1 AS port FROM playground-containers LIMIT 1")
  err = row.Scan(&ip, &port)
  
  if err != nil {
    fmt.Printf("Error get new ip & port: %v\n", err)
    return false
  }

  container_id, err := getOutput(
    "docker", 
    "run", 
    "--name", fmt.Sprintf("%s-playground", username), 
    "-h", "oide",
    "--net", "custom", 
    "--ip", ip, 
    "-m", "100m",
    "--cpus", "0.05",
    "-p", fmt.Sprintf("%d:22", port), 
    "-d", "oide-playground",
  )
  if err != nil {
    fmt.Printf("Error execute command: %v\n", err)
    return false
  }

  container_storage, err := getOutput("docker", "inspect", "--format={{.GraphDriver.Data.MergedDir}}", container_id[:len(container_id)-1])
  if err != nil {
    fmt.Printf("Error inspect container: %v\n", err)
    return false
  }
  fmt.Printf("ID: %s\nPath: %s\n", container_id[:len(container_id)-1], container_storage[:len(container_storage)-1])

  _, err = tx.Exec(
    "INSERT INTO playground-containers (user_id, container_id, container_ip, container_storage, ssh_port) VALUES ((SELECT id FROM users WHERE username = $1), $2, $3, $4, $5)", 
    username, container_id[:len(container_id)-1], ip, container_storage[:len(container_storage)-1], port,
  )
  if err != nil {
    fmt.Printf("Error insert into containers: %v\n", err)
    tx.Rollback()
    return false
  }
  tx.Commit()
  return true
}

func CreateDeploymentContainer(db *sql.DB, username string) bool {
  var ip string
  
  tx, err := db.Begin()
  if err != nil {
    fmt.Printf("Error start transaction: %v\n", err)
    return false
  }

  row := tx.QueryRow("SELECT MAX(container_ip)::inet+1 AS ip FROM deployment-containers LIMIT 1")
  err = row.Scan(&ip)
  
  if err != nil {
    fmt.Printf("Error get new ip & port: %v\n", err)
    return false
  }

  container_id, err := getOutput(
    "docker", 
    "run", 
    "--name", fmt.Sprintf("%s-deployment", username), 
    "-h", "oide",
    "--net", "custom", 
    "--ip", ip, 
    "-m", "100m",
    "--cpus", "0.05",
    "-d", "oide-deployment",
  )
  if err != nil {
    fmt.Printf("Error execute command: %v\n", err)
    return false
  }

  container_storage, err := getOutput("docker", "inspect", "--format={{.GraphDriver.Data.MergedDir}}", container_id[:len(container_id)-1])
  if err != nil {
    fmt.Printf("Error inspect container: %v\n", err)
    return false
  }
  fmt.Printf("ID: %s\nPath: %s\n", container_id[:len(container_id)-1], container_storage[:len(container_storage)-1])

  _, err = tx.Exec(
    "INSERT INTO deployment-containers (user_id, container_id, container_ip, container_storage) VALUES ((SELECT id FROM users WHERE username = $1), $2, $3, $4)", 
    username, container_id[:len(container_id)-1], ip, container_storage[:len(container_storage)-1],
  )
  if err != nil {
    fmt.Printf("Error insert into containers: %v\n", err)
    tx.Rollback()
    return false
  }
  tx.Commit()
  return true
}

package data

import (
    "context"
    "database/sql"
    "slices"
    "time"
    "github.com/lib/pq"
)
type Permissions []string

func (p Permissions) Include(code string) bool {
    return slices.Contains(p, code)
}

type PermissionModel struct {
    DB *sql.DB
}

func (p PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
    query := `
               SELECT permissions.code
               FROM permissions 
               INNER JOIN users_permissions ON 
               users_permissions.permission_id = permissions.id
			   INNER JOIN users ON users_permissions.user_id = users.id
			   WHERE users.id = $1
			`
	  ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	  defer cancel()
  
	  rows, err := p.DB.QueryContext(ctx, query, userID)
	  if err != nil {
		  return nil, err
	  }  
    defer rows.Close()
    var permissions Permissions
    for rows.Next() {
        var permission string

        err := rows.Scan(&permission)
        if err != nil {
            return nil, err
        }
		permissions = append(permissions, permission)
		} 
	
		err = rows.Err()
		if err != nil {
			return nil, err
		}
	
	 return permissions, nil
	
	}
	
func (p PermissionModel) AddForUser(userID int64, codes ...string) error {
		query := `
        INSERT INTO users_permissions
        SELECT $1, permissions.id FROM permissions 
        WHERE permissions.code = ANY($2)
       `
   ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
   defer cancel()

   _, err := p.DB.ExecContext(ctx, query, userID, pq.Array(codes))

   return err
}
	  
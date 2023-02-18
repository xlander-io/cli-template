package user_mgr

import (
	"errors"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/redis_plugin"
	"github.com/coreservice-io/cli-template/src/common/json"
	"github.com/coreservice-io/cli-template/src/common/smart_cache"
	"github.com/coreservice-io/cli-template/src/common/token_mgr"
	"github.com/coreservice-io/utils/hash_util"
	"gorm.io/gorm"
)

func RolesToStr(roles ...string) string {
	r_str, _ := json.Marshal(roles)
	return string(r_str)
}

func PermissionsToStr(permissions ...string) string {
	p_str, _ := json.Marshal(permissions)
	return string(p_str)
}

func GenRandUserToken(isSuperUser bool) string {
	if isSuperUser {
		return token_mgr.TokenMgr.GenSuperToken()
	} else {
		return token_mgr.TokenMgr.GenToken()
	}
}

func CreateUser(tx *gorm.DB, email string, passwd string, isSuperUser bool, roles []string, permissions []string, ipv4 string) (*UserModel, error) {
	sha256_passwd := hash_util.SHA256String(passwd)
	token := GenRandUserToken(isSuperUser)

	// check roles
	if !RolesDefined(roles) {
		return nil, errors.New("roles not defined")
	}

	// check permissions
	if !PermissionsDefined(permissions) {
		return nil, errors.New("permissions not defined")
	}

	user := &UserModel{
		Email:           email,
		Password:        sha256_passwd,
		Token:           token,
		Roles:           RolesToStr(roles...),
		Permissions:     PermissionsToStr(permissions...),
		Roles_map:       make(map[string]string),
		Permissions_map: make(map[string]string),
		Forbidden:       false,
		Register_ipv4:   ipv4,
	}
	for _, role := range roles {
		user.Roles_map[role] = role
	}

	for _, p := range permissions {
		user.Permissions_map[p] = p
	}

	if err := tx.Table(TABLE_NAME_USER).Create(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateUser(tx *gorm.DB, updateData map[string]interface{}, id int64) error {
	queryResult, err := QueryUser(tx, &id, nil, nil, nil, nil, 1, 0, false, false)
	if err != nil {
		basic.Logger.Errorln("UpdateNodeUser queryUsers error:", err, "id:", id)
		return err
	}
	if len(queryResult.Users) == 0 {
		return errors.New("user not exist")
	}

	update_result := tx.Table(TABLE_NAME_USER).Where("id =?", id).Updates(updateData)
	if update_result.Error != nil {
		return update_result.Error
	}

	if update_result.RowsAffected == 0 {
		return errors.New("0 raw affected")
	}

	// update cache , for fast api middleware token auth
	QueryUser(tx, nil, &queryResult.Users[0].Token, nil, nil, nil, 1, 0, false, true)

	return nil
}

type QueryUserResult struct {
	Users       []*UserModel
	Total_count int64
}

func QueryUser(tx *gorm.DB, id *int64, token *string, emailPattern *string, email *string, forbidden *bool, limit int, offset int, fromCache bool, updateCache bool) (*QueryUserResult, error) {

	if emailPattern != nil && email != nil {
		return &QueryUserResult{
			Users:       []*UserModel{},
			Total_count: 0,
		}, errors.New("emailPattern ,email :can't be set at the same time")
	}

	// gen_key
	ck := smart_cache.NewConnectKey("users")
	ck.C_Str_Ptr("token", token).
		C_Str_Ptr("emailPattern", emailPattern).
		C_Str_Ptr("email", email).
		C_Bool_Ptr("forbidden", forbidden).
		C_Int64_Ptr("id", id).
		C_Int(limit).
		C_Int(offset)

	key := redis_plugin.GetInstance().GenKey(ck.String())

	// ///
	resultHolderAlloc := func() interface{} {
		return &QueryUserResult{
			Users:       []*UserModel{},
			Total_count: 0,
		}
	}

	SlowQueryDefaultTTL := func() *smart_cache.QueryCacheTTL {
		return &smart_cache.QueryCacheTTL{
			Redis_ttl_secs: 300,
			Ref_ttl_secs:   5,
		}
	}

	query := func(resultHolder interface{}) (*smart_cache.QueryCacheTTL, error) {
		queryResult := resultHolder.(*QueryUserResult)

		query := tx.Table(TABLE_NAME_USER)
		if id != nil {
			query.Where("id = ?", *id)
		}

		if token != nil {
			query.Where("token = ?", *token)
		}

		if emailPattern != nil {
			query.Where("email LIKE ?", "%"+*emailPattern+"%")
		}

		if email != nil {
			query.Where("email = ?", *email)
		}

		if forbidden != nil {
			query.Where("forbidden = ?", *forbidden)
		}

		query.Count(&queryResult.Total_count)
		if limit > 0 {
			query.Limit(limit)
		}
		if offset > 0 {
			query.Offset(offset)
		}

		err := query.Find(&queryResult.Users).Error
		if err != nil {
			return smart_cache.SlowQueryTTL_ZERO, err
		}

		// equip the related info
		for _, user := range queryResult.Users {

			user.Roles_map = make(map[string]string)
			var userRoles []string
			json.Unmarshal([]byte(user.Roles), &userRoles)
			for _, u_r := range userRoles {
				user.Roles_map[u_r] = u_r
			}

			user.Permissions_map = make(map[string]string)
			var userPermissions []string
			json.Unmarshal([]byte(user.Permissions), &userPermissions)
			for _, u_p := range userPermissions {
				user.Permissions_map[u_p] = u_p
			}
		}

		if len(queryResult.Users) == 0 {
			return smart_cache.SlowQueryTTL_NOT_FOUND, nil // if no record, cache 30 secs in redis and 5 secs in ref
		} else {
			return SlowQueryDefaultTTL(), nil // if len(record)>0, cache 300 secs in redis and 5 secs in ref
		}

	}

	s_query := &smart_cache.SlowQuery{
		DefaultTTL: SlowQueryDefaultTTL,
		Query:      query,
	}

	//
	sq_result, sq_err := smart_cache.SmartQueryCacheSlow(key, resultHolderAlloc, true, fromCache, updateCache, s_query, "QueryUser")

	//
	if sq_err != nil {
		return nil, sq_err
	} else {
		return sq_result.(*QueryUserResult), nil
	}
}

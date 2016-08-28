package models

import (
	"errors"
	"log"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
)

//权限模块表
type Privilege struct {
	Id     int64
	Title  string `orm:"size(100)" form:"Title"  valid:"Required"`
	Name   string `orm:"size(100)" form:"Name"  valid:"Required"`
	Remark string `orm:"null;size(200)" form:"Remark" valid:"MaxSize(200)"`
	Status int    `orm:"default(2)" form:"Status" valid:"Range(1,2)"`
}

func (r *Privilege) TableName() string {
	return beego.AppConfig.String("rbac_privilege_table")
}

func init() {
	//orm.RegisterModel(new(Privilege))
}

func checkPri(g *Privilege) (err error) {
	valid := validation.Validation{}
	b, _ := valid.Valid(&g)
	if !b {
		for _, err := range valid.Errors {
			log.Println(err.Key, err.Message)
			return errors.New(err.Message)
		}
	}
	return nil
}

//get role list
func GetPrilist(page int64, page_size int64, sort string) (roles []orm.Params, count int64) {
	o := orm.NewOrm()
	role := new(Privilege)
	qs := o.QueryTable(role)
	var offset int64
	if page <= 1 {
		offset = 0
	} else {
		offset = (page - 1) * page_size
	}
	qs.Limit(page_size, offset).OrderBy(sort).Values(&roles)
	count, _ = qs.Count()
	return roles, count
}

func AddPri(r *Privilege) (int64, error) {
	if err := checkPri(r); err != nil {
		return 0, err
	}
	o := orm.NewOrm()
	pri := new(Privilege)
	pri.Title = r.Title
	pri.Name = r.Name
	pri.Remark = r.Remark
	pri.Status = r.Status

	id, err := o.Insert(pri)
	return id, err
}

func UpdatePri(r *Privilege) (int64, error) {
	if err := checkPri(r); err != nil {
		return 0, err
	}
	o := orm.NewOrm()
	role := make(orm.Params)
	if len(r.Title) > 0 {
		role["Title"] = r.Title
	}
	if len(r.Name) > 0 {
		role["Name"] = r.Name
	}
	if len(r.Remark) > 0 {
		role["Remark"] = r.Remark
	}
	if r.Status != 0 {
		role["Status"] = r.Status
	}
	if len(role) == 0 {
		return 0, errors.New("update field is empty")
	}
	var table Privilege
	num, err := o.QueryTable(table).Filter("Id", r.Id).Update(role)
	return num, err
}

func DelPriById(Id int64) (int64, error) {
	o := orm.NewOrm()
	status, err := o.Delete(&Privilege{Id: Id})
	return status, err
}

func GetNodelistByPriId(Id int64) (nodes []orm.Params, count int64) {
	o := orm.NewOrm()
	node := new(Privilege)
	count, _ = o.QueryTable(node).Filter("Role__Role__Id", Id).Values(&nodes)
	return nodes, count
}

func DelGroupPri(roleid int64, groupid int64) error {
	var nodes []*Privilege
	var node Privilege
	role := Role{Id: roleid}
	o := orm.NewOrm()
	num, err := o.QueryTable(node).Filter("Group", groupid).RelatedSel().All(&nodes)
	if err != nil {
		return err
	}
	if num < 1 {
		return nil
	}
	for _, n := range nodes {
		m2m := o.QueryM2M(n, "Privilege")
		_, err1 := m2m.Remove(&role)
		if err1 != nil {
			return err1
		}
	}
	return nil
}
func AddRolePri(roleid int64, nodeid int64) (int64, error) {
	o := orm.NewOrm()
	role := Privilege{Id: roleid}
	node := Node{Id: nodeid}
	m2m := o.QueryM2M(&node, "Privilege")
	num, err := m2m.Add(&role)
	return num, err
}

func DelUserPri(roleid int64) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("user_roles").Filter("role_id", roleid).Delete()
	return err
}
func AddPriUser(roleid int64, userid int64) (int64, error) {
	o := orm.NewOrm()
	role := Privilege{Id: roleid}
	user := User{Id: userid}
	m2m := o.QueryM2M(&user, "Privilege")
	num, err := m2m.Add(&role)
	return num, err
}

func GetUserByPriId(roleid int64) (users []orm.Params, count int64) {
	o := orm.NewOrm()
	user := new(User)
	count, _ = o.QueryTable(user).Filter("Role__Role__Id", roleid).Values(&users)
	return users, count
}

func AccessListPri(uid int64) (list []orm.Params, err error) {
	var roles []orm.Params
	o := orm.NewOrm()
	role := new(Role)
	_, err = o.QueryTable(role).Filter("User__User__Id", uid).Values(&roles)
	if err != nil {
		return nil, err
	}
	var nodes []orm.Params
	node := new(Node)
	for _, r := range roles {
		_, err := o.QueryTable(node).Filter("Role__Role__Id", r["Id"]).Values(&nodes)
		if err != nil {
			return nil, err
		}
		for _, n := range nodes {
			list = append(list, n)
		}
	}
	return list, nil
}

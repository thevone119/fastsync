/**********************************************
** @Des: This file ...
** @Author: haodaquan
** @Date:   2017-09-16 15:42:43
** @Last Modified by:   haodaquan
** @Last Modified time: 2017-09-17 11:48:17
***********************************************/
package models

import (
	"errors"
)

type Admin struct {
	Id         int
	LoginName  string
	RealName   string
	Password   string
	RoleIds    string
	Phone      string
	Email      string
	Salt       string
	LastLogin  int64
	LastIp     string
	Status     int
	CreateId   int
	UpdateId   int
	CreateTime int64
	UpdateTime int64
}

func (a *Admin) TableName() string {
	return TableName("uc_admin")
}


func AdminGetByName(loginName string) (*Admin, error) {
	a := new(Admin)
	if loginName=="admin"{
		a.Id=1
		a.RealName="admin"
		a.Password="admin"
		a.Status=1
		return a,nil
	}
	return nil, errors.New("not found")
}



func AdminGetById(id int) (*Admin, error) {
	r := new(Admin)
	if id==1{
		r.LoginName="admin"
		r.Id=1
		r.RealName="admin"
		r.Password="admin"
		r.Status=1
		return r, nil
	}


	return nil, errors.New("not found")
}

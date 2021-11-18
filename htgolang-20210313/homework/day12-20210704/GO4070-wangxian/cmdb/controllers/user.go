package controllers

import (
	"cmdb/forms"
	"cmdb/models"
	"cmdb/services"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/beego/beego/v2/adapter/validation"
	"github.com/beego/beego/v2/server/web"
)

type UserController struct {
	AuthenticationController
}

func (c *UserController) Query() {

	keyword := c.GetString("q")
	users := services.QueryUser(keyword)
	// fmt.Println(users)
	c.Data["Users"] = users
	c.TplName = "user/list.html"

}

func (c *UserController) Delete() {
	id, _ := c.GetInt64("id")
	services.DeleteUser(id)
	c.Redirect(web.URLFor("UserController.Query"), http.StatusFound)
}

func (c *UserController) Add() {

	var form forms.AddUserForm
	var valid validation.Validation
	roles := services.QueryRole()
	if c.Ctx.Input.IsPost() {
		err := c.ParseForm(&form)
		// fmt.Println(form)
		if err == nil {
			if b, err := valid.Valid(&form); err != nil {
				log.Println("valid add user form data error.", err)
				valid.SetError("user", "验证数据错误")
			} else if b {
				err := services.AddUser(form.Username, form.Sex, form.Tel, form.Email, form.Addr, form.Description, form.Password, form.Birthday, form.RoleId)
				if err == nil {
					c.Redirect(web.URLFor("UserController.Query"), http.StatusFound)
					return
				}
				log.Println("add user faild.", err)
				valid.SetError("user", "添加用户失败")
			}

		} else {
			// fmt.Println(err)
			valid.SetError("user", "提交数据错误")
		}
	}

	// fmt.Println(valid.ErrorsMap)
	c.Data["Roles"] = roles
	c.Data["ErrMsgs"] = valid.ErrorMap()
	c.TplName = "user/add.html"
}

func (c *UserController) Edit() {
	var form forms.EditUserForm
	var valid validation.Validation
	var user *models.User = &models.User{}
	roles := services.QueryRole()

	if c.Ctx.Input.IsGet() {
		id, _ := c.GetInt64("id")
		user = services.QueryUserByID(id)
	}

	if c.Ctx.Input.IsPost() {
		err := c.ParseForm(&form)
		// fmt.Printf("%#v", form)
		if err == nil {
			if b, err := valid.Valid(&form); err != nil {
				log.Println("valid edit user form data error.", err)
				valid.SetError("user", "验证数据错误")
			} else if b {
				//更新数据
				err := services.EditUser(form.Id, form.Username, form.Sex, form.Birthday, form.Tel, form.Email, form.Addr, form.Description, form.RoleId)
				if err == nil {
					c.Redirect(web.URLFor("UserController.Query"), http.StatusFound)
					return
				}
				log.Println("edit user info faild.", err)
				valid.SetError("user", "修改用户信息失败")
			} else {
				//验证数据存在错误时，模板上回显错误信息
				bir, _ := time.Parse("2006-01-02", form.Birthday)
				user.Id = form.Id
				user.Name = form.Username
				user.Sex = func() bool { return form.Sex != "0" }()
				user.Birthday = &bir
				user.Telephone = form.Tel
				user.Email = form.Email
				user.Addr = form.Addr
				user.Description = form.Description
				user.RoleId = form.RoleId
			}
		} else {
			valid.SetError("user", "提交数据错误")
		}
	}

	// fmt.Println(valid.ErrorMap())
	c.Data["User"] = user
	c.Data["Roles"] = roles
	c.Data["ErrMsgs"] = valid.ErrorMap()
	c.TplName = "user/edit.html"
}

func (c *UserController) ModifyPw() {
	var valid validation.Validation
	var form forms.ModifyPwForm
	var success string

	if c.Ctx.Input.IsPost() {
		//获取当前用户
		form.User = c.CurrentUser

		err := c.ParseForm(&form)

		if err == nil {
			fmt.Println("valid data")
			if b, err := valid.Valid(&form); err != nil {
				log.Println("valid modifypw data error.", err)
				valid.SetError("userpassword", "验证数据错误")
			} else if b {
				form.User.SetPassword(form.NewPassword)
				err := services.ModifyPw(form.User.Id, form.User.Password)
				if err == nil {
					//修改密码成功，销毁session,重新登录
					c.DestroySession()
					success = "修改密码成功，请刷新页面重新登录"
				} else {
					//修改失败
					log.Println("modify password error.", err)
					valid.SetError("userpassword", "修改密码失败")
				}
			}
		} else {
			log.Println("parse modifypw data error.", err)
			valid.SetError("userpassword", "提交数据错误")
		}
	}

	c.Data["Success"] = success
	c.Data["ErrMsgs"] = valid.ErrorMap()
	c.TplName = "user/changepw.html"
}

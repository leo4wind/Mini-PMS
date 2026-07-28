package main

import (
	"fmt"
	"time"

	"minipms/internal/config"
	"minipms/internal/database"
	"minipms/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func ptr[T any](v T) *T { return &v }

func mustHash(pwd string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func date(y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	return &t
}

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		panic(err)
	}

	// 清空业务数据，保留角色/菜单与 admin（id=1）
	if err := db.Exec("SET FOREIGN_KEY_CHECKS=0").Error; err != nil {
		panic(err)
	}
	for _, t := range []string{"attachment", "sprint_story", "bug", "sprint", "story", "project", "product"} {
		if err := db.Exec("TRUNCATE TABLE `" + t + "`").Error; err != nil {
			panic(err)
		}
	}
	// 删掉除 admin 外的用户及关联
	db.Exec("DELETE FROM user_role WHERE user_id <> 1")
	db.Exec("DELETE FROM `user` WHERE id <> 1")
	db.Exec("ALTER TABLE `user` AUTO_INCREMENT = 2")
	if err := db.Exec("SET FOREIGN_KEY_CHECKS=1").Error; err != nil {
		panic(err)
	}

	hash := mustHash("password")
	now := time.Now()

	users := []model.User{
		{Account: "zhangsan", PasswordHash: hash, Realname: "张三", Email: ptr("zhangsan@example.com"), Status: "active"},
		{Account: "lisi", PasswordHash: hash, Realname: "李四", Email: ptr("lisi@example.com"), Status: "active"},
		{Account: "wangwu", PasswordHash: hash, Realname: "王五", Email: ptr("wangwu@example.com"), Status: "active"},
		{Account: "zhaoliu", PasswordHash: hash, Realname: "赵六", Email: ptr("zhaoliu@example.com"), Status: "active"},
		{Account: "chenqi", PasswordHash: hash, Realname: "陈七", Email: ptr("chenqi@example.com"), Status: "active"},
		{Account: "sunba", PasswordHash: hash, Realname: "孙八", Email: ptr("sunba@example.com"), Status: "active"},
		{Account: "zhoujiu", PasswordHash: hash, Realname: "周九", Email: ptr("zhoujiu@example.com"), Status: "active"},
		{Account: "wushi", PasswordHash: hash, Realname: "吴十", Email: ptr("wushi@example.com"), Status: "active"},
		{Account: "zhengyi", PasswordHash: hash, Realname: "郑一", Email: ptr("zhengyi@example.com"), Status: "active"},
		{Account: "fenger", PasswordHash: hash, Realname: "冯二", Email: ptr("fenger@example.com"), Status: "active"},
		{Account: "chusan", PasswordHash: hash, Realname: "楚三", Email: ptr("chusan@example.com"), Status: "active"},
		{Account: "weisi", PasswordHash: hash, Realname: "魏四", Email: ptr("weisi@example.com"), Status: "disabled"},
	}
	if err := db.Create(&users).Error; err != nil {
		panic(err)
	}
	// 角色：2产品 1开发 3测试 4经理
	roleMap := []uint64{2, 1, 3, 4, 1, 2, 3, 1, 2, 3, 1, 1}
	for i, u := range users {
		db.Create(&model.UserRole{UserID: u.ID, RoleID: roleMap[i], CreatedAt: now})
	}

	productNames := []string{
		"进销存系统", "客户关系管理", "在线商城", "仓储管理", "财务结算",
		"人力资源", "办公协同", "移动考勤", "数据报表", "开放平台",
		"运维监控", "知识库",
	}
	productCodes := []string{"erp", "crm", "mall", "wms", "fms", "hr", "oa", "att", "bi", "open", "ops", "kb"}
	products := make([]model.Product, 0, len(productNames))
	for i, name := range productNames {
		status := "normal"
		if i >= 10 {
			status = "closed"
		}
		po := users[i%len(users)].ID
		p := model.Product{
			Name: name, Code: ptr(productCodes[i]), Status: status,
			PO: &po, Description: ptr(name + "演示产品说明"), CreatedBy: 1,
		}
		products = append(products, p)
	}
	if err := db.Create(&products).Error; err != nil {
		panic(err)
	}

	// 每个产品至少 1 个默认项目，再补到约 12+ 项目
	projects := make([]model.Project, 0, 16)
	for i, p := range products {
		code := productCodes[i] + "-1.0"
		pm := users[i%len(users)].ID
		st := "wait"
		if i%4 == 1 {
			st = "doing"
		} else if i%4 == 2 {
			st = "suspended"
		} else if i%4 == 3 {
			st = "closed"
		}
		projects = append(projects, model.Project{
			ProductID: p.ID, Name: p.Name + "1.0", Code: &code, Status: st,
			Begin: date(2026, 1, 1), End: date(2026, 6, 30), PM: &pm,
			Description: ptr("自动演示项目"), CreatedBy: 1,
		})
	}
	// 额外项目：前 4 个产品各再加一个 2.0
	for i := 0; i < 4; i++ {
		p := products[i]
		code := productCodes[i] + "-2.0"
		pm := users[(i+3)%len(users)].ID
		projects = append(projects, model.Project{
			ProductID: p.ID, Name: p.Name + "2.0", Code: &code, Status: "doing",
			Begin: date(2026, 7, 1), End: date(2026, 12, 31), PM: &pm,
			Description: ptr("二期演示项目"), CreatedBy: 1,
		})
	}
	if err := db.Create(&projects).Error; err != nil {
		panic(err)
	}

	storyTitles := []string{
		"用户登录与鉴权", "商品列表与搜索", "购物车结算", "订单状态流转",
		"库存出入库", "供应商管理", "客户档案", "合同审批流程",
		"报表导出", "消息通知中心", "角色权限配置", "附件上传下载",
		"支付对接", "发票开具", "移动端首页", "优惠券规则",
		"退货退款流程", "物流轨迹同步", "会员积分", "审计日志查询",
	}
	stories := make([]model.Story, 0, len(storyTitles))
	for i, title := range storyTitles {
		p := products[i%12] // 12 个产品轮询
		typ := "story"
		st := "active"
		switch i % 6 {
		case 0:
			typ = "planning"
			st = "draft"
		case 5:
			st = "closed"
		}
		assignee := users[i%len(users)].ID
		est := float64(2 + i%8)
		stories = append(stories, model.Story{
			ProductID: p.ID, Type: typ, Title: title,
			Description: ptr("演示需求：" + title), Pri: uint8(1 + i%4),
			Status: st, Estimate: &est, AssignedTo: &assignee, OpenedBy: 1,
		})
	}
	if err := db.Create(&stories).Error; err != nil {
		panic(err)
	}

	sprintNames := []string{
		"一月冲刺", "二月冲刺", "三月冲刺", "四月冲刺", "五月冲刺", "六月冲刺",
		"七月冲刺", "八月冲刺", "九月冲刺", "十月冲刺", "十一月冲刺", "十二月冲刺",
	}
	sprints := make([]model.Sprint, 0, len(sprintNames))
	for i, name := range sprintNames {
		proj := projects[i%len(projects)]
		st := "wait"
		if i%4 == 1 {
			st = "doing"
		} else if i%4 == 2 {
			st = "done"
		} else if i%4 == 3 {
			st = "closed"
		}
		m := time.Month(i + 1)
		sprints = append(sprints, model.Sprint{
			ProjectID: proj.ID, Name: name, Status: st,
			Begin: date(2026, m, 1), End: date(2026, m, 28),
			Goal: ptr(name + "目标：交付核心功能"),
		})
	}
	if err := db.Create(&sprints).Error; err != nil {
		panic(err)
	}

	// 关联：同产品需求拉入迭代（doing 仅挂 type=story 且 active）
	for _, sp := range sprints {
		var proj model.Project
		db.First(&proj, sp.ProjectID)
		linked := 0
		for _, st := range stories {
			if st.ProductID != proj.ProductID {
				continue
			}
			if sp.Status == "doing" && (st.Type != "story" || st.Status != "active") {
				continue
			}
			if st.Type != "story" && st.Status == "draft" {
				continue
			}
			if err := db.Create(&model.SprintStory{
				ProjectID: proj.ID, SprintID: sp.ID, ProductID: proj.ProductID, StoryID: st.ID,
			}).Error; err != nil {
				continue
			}
			linked++
			if linked >= 2 {
				break
			}
		}
	}

	bugTitles := []string{
		"登录按钮点击无响应", "列表分页总数错误", "导出文件乱码", "上传超大文件失败",
		"权限校验绕过", "搜索关键词为空报错", "日期时区显示偏差", "重复提交产生两单",
		"移动端布局错位", "通知未已读标记", "附件预览白屏", "关闭后仍可编辑",
		"严重级别筛选无效", "指派人下拉空白",
	}
	bugs := make([]model.Bug, 0, len(bugTitles))
	for i, title := range bugTitles {
		proj := projects[i%len(projects)]
		var prod model.Product
		for _, pp := range products {
			if pp.ID == proj.ProductID {
				prod = pp
				break
			}
		}
		sp := sprints[i%len(sprints)]
		// 迭代尽量挂到同一项目：找不到则只用项目
		var sprintID *uint64
		for _, s := range sprints {
			if s.ProjectID == proj.ID {
				id := s.ID
				sprintID = &id
				break
			}
		}
		if sprintID == nil && i%2 == 0 {
			// 回退：改用该 sprint 所属项目，保持级联一致
			proj = projects[0]
			for _, p := range projects {
				if p.ID == sp.ProjectID {
					proj = p
					break
				}
			}
			for _, pp := range products {
				if pp.ID == proj.ProductID {
					prod = pp
					break
				}
			}
			id := sp.ID
			sprintID = &id
		}
		var storyID *uint64
		for _, st := range stories {
			if st.ProductID == prod.ID {
				id := st.ID
				storyID = &id
				break
			}
		}
		status := "active"
		var resolution *string
		var resolvedBy *uint64
		if i%4 == 1 {
			status = "resolved"
			resolution = ptr("fixed")
			rb := users[i%len(users)].ID
			resolvedBy = &rb
		} else if i%4 == 2 {
			status = "closed"
			resolution = ptr("bydesign")
			rb := users[i%len(users)].ID
			resolvedBy = &rb
		}
		assignee := users[(i+1)%len(users)].ID
		projectID := proj.ID
		bugs = append(bugs, model.Bug{
			ProductID: prod.ID, ProjectID: &projectID, SprintID: sprintID, StoryID: storyID,
			Title: title, Steps: ptr("1.打开页面\n2.复现问题\n3.观察结果"),
			Severity: uint8(1 + i%4), Pri: uint8(1 + (i+1)%4), Status: status,
			Resolution: resolution, AssignedTo: &assignee, OpenedBy: 1, ResolvedBy: resolvedBy,
		})
	}
	if err := db.Create(&bugs).Error; err != nil {
		panic(err)
	}

	printCount(db)
	fmt.Println("demo seed ok. 账号密码均为 password（admin / zhangsan 等）")
}

func printCount(db *gorm.DB) {
	items := []struct {
		label string
		model interface{}
		where string
	}{
		{"user", &model.User{}, "deleted = 0"},
		{"product", &model.Product{}, "deleted = 0"},
		{"project", &model.Project{}, "deleted = 0"},
		{"story", &model.Story{}, "deleted = 0"},
		{"sprint", &model.Sprint{}, "deleted = 0"},
		{"sprint_story", &model.SprintStory{}, "1=1"},
		{"bug", &model.Bug{}, "deleted = 0"},
	}
	for _, it := range items {
		var n int64
		q := db.Model(it.model)
		if it.where != "" {
			q = q.Where(it.where)
		}
		q.Count(&n)
		fmt.Printf("%s: %d\n", it.label, n)
	}
}

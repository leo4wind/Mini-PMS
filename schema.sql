-- MiniPMS schema (MySQL 8+)
-- Model: Product 1-N Project 1-N Sprint
--         Story/Bug belong to Product; Story pulled into Sprint; Bug optionally linked
--         Attachment polymorphic for story/bug
--         No Task in V1
--         Creating a product auto-creates project "{product.name}1.0" (app rule)
-- Charset: utf8mb4 | Engine: InnoDB

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE DATABASE IF NOT EXISTS `minipms`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE `minipms`;

-- ---------------------------------------------------------------------------
-- Drop all (order-safe)
-- ---------------------------------------------------------------------------

DROP TABLE IF EXISTS `role_menu`;
DROP TABLE IF EXISTS `user_role`;
DROP TABLE IF EXISTS `attachment`;
DROP TABLE IF EXISTS `sprint_story`;
DROP TABLE IF EXISTS `bug`;
DROP TABLE IF EXISTS `sprint`;
DROP TABLE IF EXISTS `story`;
DROP TABLE IF EXISTS `project`;
DROP TABLE IF EXISTS `product`;
DROP TABLE IF EXISTS `menu`;
DROP TABLE IF EXISTS `role`;
DROP TABLE IF EXISTS `user`;

-- ---------------------------------------------------------------------------
-- Auth / RBAC
-- ---------------------------------------------------------------------------

CREATE TABLE `user` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `account`       VARCHAR(64)  NOT NULL COMMENT '登录名',
  `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `realname`      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '姓名',
  `email`         VARCHAR(128) NULL DEFAULT NULL,
  `status`        ENUM('active','disabled') NOT NULL DEFAULT 'active',
  `created_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted`       TINYINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_account` (`account`),
  KEY `idx_user_status` (`status`),
  KEY `idx_user_deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户';

CREATE TABLE `role` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`       VARCHAR(32)  NOT NULL COMMENT 'dev/product/qa/manager',
  `name`       VARCHAR(64)  NOT NULL,
  `builtin`    TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '1=内置不可删',
  `remark`     VARCHAR(255) NULL DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色';

CREATE TABLE `menu` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `parent_id`  BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '父菜单，NULL=顶级',
  `code`       VARCHAR(64)  NOT NULL COMMENT '唯一编码，如 product.list',
  `name`       VARCHAR(64)  NOT NULL COMMENT '显示名',
  `type`       ENUM('dir','menu','button') NOT NULL DEFAULT 'menu',
  `path`       VARCHAR(255) NULL DEFAULT NULL COMMENT '前端路由，button 可空',
  `icon`       VARCHAR(64)  NULL DEFAULT NULL,
  `sort`       INT NOT NULL DEFAULT 0,
  `status`     ENUM('enabled','disabled') NOT NULL DEFAULT 'enabled',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_menu_code` (`code`),
  KEY `idx_menu_parent` (`parent_id`),
  KEY `idx_menu_status` (`status`),
  CONSTRAINT `fk_menu_parent` FOREIGN KEY (`parent_id`) REFERENCES `menu` (`id`)
    ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单';

CREATE TABLE `user_role` (
  `user_id`    BIGINT UNSIGNED NOT NULL,
  `role_id`    BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`, `role_id`),
  KEY `idx_user_role_role` (`role_id`),
  CONSTRAINT `fk_user_role_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_user_role_role` FOREIGN KEY (`role_id`) REFERENCES `role` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-角色';

CREATE TABLE `role_menu` (
  `role_id`    BIGINT UNSIGNED NOT NULL,
  `menu_id`    BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`role_id`, `menu_id`),
  KEY `idx_role_menu_menu` (`menu_id`),
  CONSTRAINT `fk_role_menu_role` FOREIGN KEY (`role_id`) REFERENCES `role` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_role_menu_menu` FOREIGN KEY (`menu_id`) REFERENCES `menu` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色-菜单';

-- ---------------------------------------------------------------------------
-- Product / Story / Project / Sprint / Bug
-- ---------------------------------------------------------------------------

CREATE TABLE `product` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`        VARCHAR(110) NOT NULL,
  `code`        VARCHAR(45)  NULL DEFAULT NULL,
  `status`      ENUM('normal','closed') NOT NULL DEFAULT 'normal',
  `po`          BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '产品负责人 user_id',
  `description` TEXT NULL,
  `created_by`  BIGINT UNSIGNED NOT NULL,
  `created_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted`     TINYINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_code` (`code`),
  KEY `idx_product_status` (`status`),
  KEY `idx_product_po` (`po`),
  KEY `idx_product_deleted` (`deleted`),
  CONSTRAINT `fk_product_po` FOREIGN KEY (`po`) REFERENCES `user` (`id`)
    ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_product_created_by` FOREIGN KEY (`created_by`) REFERENCES `user` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品';

CREATE TABLE `story` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `product_id`  BIGINT UNSIGNED NOT NULL COMMENT '归属产品',
  `type`        ENUM('planning','story') NOT NULL DEFAULT 'planning' COMMENT 'planning=原始/规划需求 story=可交付需求',
  `title`       VARCHAR(255) NOT NULL,
  `description` TEXT NULL,
  `pri`         TINYINT UNSIGNED NOT NULL DEFAULT 3 COMMENT '1最高 4最低',
  `status`      ENUM('draft','active','closed') NOT NULL DEFAULT 'draft',
  `estimate`    DECIMAL(10,2) UNSIGNED NULL DEFAULT NULL COMMENT '估算',
  `assigned_to` BIGINT UNSIGNED NULL DEFAULT NULL,
  `opened_by`   BIGINT UNSIGNED NOT NULL,
  `created_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted`     TINYINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_story_product` (`product_id`),
  KEY `idx_story_type` (`type`),
  KEY `idx_story_status` (`status`),
  KEY `idx_story_assigned` (`assigned_to`),
  KEY `idx_story_deleted` (`deleted`),
  CONSTRAINT `fk_story_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_story_assigned` FOREIGN KEY (`assigned_to`) REFERENCES `user` (`id`)
    ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_story_opened_by` FOREIGN KEY (`opened_by`) REFERENCES `user` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='需求(归属产品；规划或可交付)';

CREATE TABLE `project` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `product_id`  BIGINT UNSIGNED NOT NULL COMMENT '所属产品 1-N',
  `name`        VARCHAR(110) NOT NULL COMMENT '如 1.0采购',
  `code`        VARCHAR(45)  NULL DEFAULT NULL,
  `status`      ENUM('wait','doing','suspended','closed') NOT NULL DEFAULT 'wait',
  `begin`       DATE NULL DEFAULT NULL,
  `end`         DATE NULL DEFAULT NULL,
  `pm`          BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '项目经理 user_id',
  `description` TEXT NULL,
  `created_by`  BIGINT UNSIGNED NOT NULL,
  `created_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted`     TINYINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_code` (`code`),
  KEY `idx_project_product` (`product_id`),
  KEY `idx_project_status` (`status`),
  KEY `idx_project_pm` (`pm`),
  KEY `idx_project_deleted` (`deleted`),
  CONSTRAINT `fk_project_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_project_pm` FOREIGN KEY (`pm`) REFERENCES `user` (`id`)
    ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_project_created_by` FOREIGN KEY (`created_by`) REFERENCES `user` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目(归属产品，如版本/模块交付)';

CREATE TABLE `sprint` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` BIGINT UNSIGNED NOT NULL,
  `name`       VARCHAR(110) NOT NULL,
  `status`     ENUM('wait','doing','done','closed') NOT NULL DEFAULT 'wait',
  `begin`      DATE NULL DEFAULT NULL COMMENT '支持3天或2周等任意周期',
  `end`        DATE NULL DEFAULT NULL,
  `goal`       TEXT NULL COMMENT '迭代目标',
  `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建人 user_id',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted`    TINYINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_sprint_project` (`project_id`),
  KEY `idx_sprint_status` (`status`),
  KEY `idx_sprint_deleted` (`deleted`),
  CONSTRAINT `fk_sprint_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_sprint_created_by` FOREIGN KEY (`created_by`) REFERENCES `user` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='迭代(归属项目)';

CREATE TABLE `sprint_story` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `project_id` BIGINT UNSIGNED NOT NULL COMMENT '冗余=sprint.project_id，便于按项目查',
  `sprint_id`  BIGINT UNSIGNED NOT NULL,
  `product_id` BIGINT UNSIGNED NOT NULL COMMENT '冗余=story.product_id',
  `story_id`   BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sprint_story` (`sprint_id`, `story_id`),
  KEY `idx_ss_project` (`project_id`),
  KEY `idx_ss_product` (`product_id`),
  KEY `idx_ss_story` (`story_id`),
  CONSTRAINT `fk_ss_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_ss_sprint` FOREIGN KEY (`sprint_id`) REFERENCES `sprint` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_ss_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_ss_story` FOREIGN KEY (`story_id`) REFERENCES `story` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='迭代拉入需求';

CREATE TABLE `bug` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `product_id`  BIGINT UNSIGNED NOT NULL COMMENT '归属产品(必填)',
  `project_id`  BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '可选关联项目',
  `sprint_id`   BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '可选关联迭代',
  `story_id`    BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '可选关联需求',
  `title`       VARCHAR(255) NOT NULL,
  `steps`       TEXT NULL COMMENT '重现步骤',
  `severity`   TINYINT UNSIGNED NOT NULL DEFAULT 3 COMMENT '1最高 4最低',
  `pri`         TINYINT UNSIGNED NOT NULL DEFAULT 3,
  `status`      ENUM('active','resolved','closed') NOT NULL DEFAULT 'active',
  `resolution`  ENUM('fixed','duplicate','willnotfix','external','bydesign','notrepro') NULL DEFAULT NULL,
  `assigned_to` BIGINT UNSIGNED NULL DEFAULT NULL,
  `opened_by`   BIGINT UNSIGNED NOT NULL,
  `resolved_by` BIGINT UNSIGNED NULL DEFAULT NULL,
  `created_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted`     TINYINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_bug_product` (`product_id`),
  KEY `idx_bug_project` (`project_id`),
  KEY `idx_bug_sprint` (`sprint_id`),
  KEY `idx_bug_story` (`story_id`),
  KEY `idx_bug_status` (`status`),
  KEY `idx_bug_assigned` (`assigned_to`),
  KEY `idx_bug_deleted` (`deleted`),
  CONSTRAINT `fk_bug_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_bug_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`id`)
    ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_bug_sprint` FOREIGN KEY (`sprint_id`) REFERENCES `sprint` (`id`)
    ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_bug_story` FOREIGN KEY (`story_id`) REFERENCES `story` (`id`)
    ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_bug_assigned` FOREIGN KEY (`assigned_to`) REFERENCES `user` (`id`)
    ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_bug_opened_by` FOREIGN KEY (`opened_by`) REFERENCES `user` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_bug_resolved_by` FOREIGN KEY (`resolved_by`) REFERENCES `user` (`id`)
    ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='缺陷(归属产品；可关联项目/迭代/需求)';

-- ---------------------------------------------------------------------------
-- Attachment (polymorphic: story / bug)
-- ---------------------------------------------------------------------------

CREATE TABLE `attachment` (
  `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `object_type`    ENUM('story','bug') NOT NULL COMMENT '关联对象类型',
  `object_id`      BIGINT UNSIGNED NOT NULL COMMENT '关联对象 ID',
  `original_name`  VARCHAR(255) NOT NULL COMMENT '上传时文件名',
  `stored_name`    VARCHAR(255) NOT NULL COMMENT '存储文件名(唯一)',
  `storage_path`   VARCHAR(512) NOT NULL COMMENT '相对或逻辑存储路径',
  `ext`            VARCHAR(16)  NOT NULL COMMENT '扩展名小写，不含点',
  `mime_type`      VARCHAR(128) NULL DEFAULT NULL,
  `size_bytes`     BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `uploaded_by`    BIGINT UNSIGNED NOT NULL,
  `created_at`     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted`        TINYINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_attachment_stored` (`stored_name`),
  KEY `idx_attachment_object` (`object_type`, `object_id`),
  KEY `idx_attachment_uploader` (`uploaded_by`),
  KEY `idx_attachment_deleted` (`deleted`),
  CONSTRAINT `fk_attachment_uploader` FOREIGN KEY (`uploaded_by`) REFERENCES `user` (`id`)
    ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='附件(多态关联 story/bug；允许 doc/docx/txt/md/图片)';

SET FOREIGN_KEY_CHECKS = 1;

-- ===========================================================================
-- Seed: roles
-- ===========================================================================

INSERT INTO `role` (`id`, `code`, `name`, `builtin`, `remark`) VALUES
  (1, 'dev',     '开发', 1, '查看需求/项目/迭代，处理缺陷'),
  (2, 'product', '产品', 1, '产品、需求、项目规划与关联'),
  (3, 'qa',      '测试', 1, '缺陷全量，需求/迭代只读'),
  (4, 'manager', '经理', 1, '全部权限');

-- ===========================================================================
-- Seed: menus
-- ===========================================================================

INSERT INTO `menu` (`id`, `parent_id`, `code`, `name`, `type`, `path`, `icon`, `sort`, `status`) VALUES
  (1,  NULL, 'dashboard',     '工作台',   'menu', '/dashboard',    'dashboard', 10, 'enabled'),
  (2,  NULL, 'product',       '产品',     'dir',  NULL,            'product',   20, 'enabled'),
  (4,  NULL, 'qa',            '测试',     'dir',  NULL,            'bug',       40, 'enabled'),
  (5,  NULL, 'system',        '系统',     'dir',  NULL,            'setting',   90, 'enabled'),

  -- 产品目录下：产品/需求/项目/迭代（原「项目」一级目录已并入）
  (10, 2, 'product.list',     '产品列表', 'menu', '/products',     NULL, 10, 'enabled'),
  (11, 2, 'story.list',       '需求列表', 'menu', '/stories',      NULL, 20, 'enabled'),
  (20, 2, 'project.list',     '项目列表', 'menu', '/projects',     NULL, 30, 'enabled'),
  (21, 2, 'sprint.list',      '迭代列表', 'menu', '/sprints',      NULL, 40, 'enabled'),

  (30, 4, 'bug.list',         '缺陷列表', 'menu', '/bugs',         NULL, 10, 'enabled'),

  (40, 5, 'user.list',        '用户管理', 'menu', '/system/users', NULL, 10, 'enabled'),
  (41, 5, 'role.list',        '角色管理', 'menu', '/system/roles', NULL, 20, 'enabled'),
  (42, 5, 'menu.list',        '菜单管理', 'menu', '/system/menus', NULL, 30, 'enabled'),

  -- product / story buttons
  (100, 10, 'product.create', '新建产品', 'button', NULL, NULL, 10, 'enabled'),
  (101, 10, 'product.edit',   '编辑产品', 'button', NULL, NULL, 20, 'enabled'),
  (102, 10, 'product.delete', '删除产品', 'button', NULL, NULL, 30, 'enabled'),
  (110, 11, 'story.create',   '新建需求', 'button', NULL, NULL, 10, 'enabled'),
  (111, 11, 'story.edit',     '编辑需求', 'button', NULL, NULL, 20, 'enabled'),
  (112, 11, 'story.delete',   '删除需求', 'button', NULL, NULL, 30, 'enabled'),
  (113, 11, 'story.attach',   '需求附件', 'button', NULL, NULL, 40, 'enabled'),

  -- project / sprint buttons
  (120, 20, 'project.create', '新建项目', 'button', NULL, NULL, 10, 'enabled'),
  (121, 20, 'project.edit',   '编辑项目', 'button', NULL, NULL, 20, 'enabled'),
  (122, 20, 'project.delete', '删除项目', 'button', NULL, NULL, 30, 'enabled'),
  (130, 21, 'sprint.create',  '新建迭代', 'button', NULL, NULL, 10, 'enabled'),
  (131, 21, 'sprint.edit',    '编辑迭代', 'button', NULL, NULL, 20, 'enabled'),
  (132, 21, 'sprint.delete',  '删除迭代', 'button', NULL, NULL, 30, 'enabled'),
  (133, 21, 'sprint.linkStory','关联需求', 'button', NULL, NULL, 40, 'enabled'),

  -- bug buttons
  (150, 30, 'bug.create',     '新建缺陷', 'button', NULL, NULL, 10, 'enabled'),
  (151, 30, 'bug.edit',       '编辑缺陷', 'button', NULL, NULL, 20, 'enabled'),
  (152, 30, 'bug.resolve',    '解决缺陷', 'button', NULL, NULL, 30, 'enabled'),
  (153, 30, 'bug.close',      '关闭缺陷', 'button', NULL, NULL, 40, 'enabled'),
  (154, 30, 'bug.delete',     '删除缺陷', 'button', NULL, NULL, 50, 'enabled'),
  (155, 30, 'bug.attach',     '缺陷附件', 'button', NULL, NULL, 60, 'enabled'),

  -- system buttons
  (160, 40, 'user.create',    '新建用户', 'button', NULL, NULL, 10, 'enabled'),
  (161, 40, 'user.edit',      '编辑用户', 'button', NULL, NULL, 20, 'enabled'),
  (162, 40, 'user.disable',   '停用用户', 'button', NULL, NULL, 30, 'enabled'),
  (163, 40, 'user.assignRole','分配角色', 'button', NULL, NULL, 40, 'enabled'),
  (170, 41, 'role.edit',      '编辑角色', 'button', NULL, NULL, 10, 'enabled'),
  (171, 41, 'role.assignMenu','分配菜单', 'button', NULL, NULL, 20, 'enabled'),
  (180, 42, 'menu.create',    '新建菜单', 'button', NULL, NULL, 10, 'enabled'),
  (181, 42, 'menu.edit',      '编辑菜单', 'button', NULL, NULL, 20, 'enabled'),
  (182, 42, 'menu.delete',    '删除菜单', 'button', NULL, NULL, 30, 'enabled');

-- manager: all
INSERT INTO `role_menu` (`role_id`, `menu_id`)
SELECT 4, `id` FROM `menu`;

-- product
INSERT INTO `role_menu` (`role_id`, `menu_id`) VALUES
  (2, 1),
  (2, 2), (2, 10), (2, 11), (2, 20), (2, 21),
  (2, 100), (2, 101), (2, 102),
  (2, 110), (2, 111), (2, 112), (2, 113),
  (2, 120), (2, 121), (2, 130), (2, 131), (2, 133),
  (2, 4), (2, 30);

-- dev: readonly product/project/sprint/story + bug handle + bug attach
INSERT INTO `role_menu` (`role_id`, `menu_id`) VALUES
  (1, 1),
  (1, 2), (1, 10), (1, 11), (1, 20), (1, 21),
  (1, 4), (1, 30),
  (1, 151), (1, 152), (1, 153), (1, 155);

-- qa：产品列表只读（建缺陷时选产品）+ 需求/项目/迭代只读 + 缺陷全量
INSERT INTO `role_menu` (`role_id`, `menu_id`) VALUES
  (3, 1),
  (3, 2), (3, 10), (3, 11), (3, 20), (3, 21),
  (3, 4), (3, 30),
  (3, 150), (3, 151), (3, 152), (3, 153), (3, 154), (3, 155);

-- admin user (bcrypt for plaintext "123456" — change before production)
INSERT INTO `user` (`id`, `account`, `password_hash`, `realname`, `email`, `status`) VALUES
  (1, 'admin', '$2a$10$9UU2gagnUPdEwmeJLuUpmeDhDWaKX8dyPW5WLCs58yuKHAEZP32JC', '系统管理员', 'admin@example.com', 'active');

INSERT INTO `user_role` (`user_id`, `role_id`) VALUES (1, 4);

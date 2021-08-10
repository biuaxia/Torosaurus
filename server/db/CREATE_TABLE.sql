create table `tbl_file` (
	`id` int(11) not null AUTO_INCREMENT, 
	`file_sha1` char(40) not null default '' comment '文件hash', 
	`file_name` varchar(256) not null default '' comment '文件名', 
	`file_size` bigint(20) default '0' comment '文件大小', 
	`file_addr` varchar(1024) not null default '' comment '文件存储位置', 
	`create_at` datetime default NOW() comment '创建日期', 
	`update_at` datetime default NOW() on update CURRENT_TIMESTAMP comment '更新日期', 
	`status` int(11) not null default '0' comment '状态（可用/禁用/已删除等状态）', 
	`ext1` int(11) default '0' comment '备用字段1', 
	`ext2` text comment '备用字段2', 
	PRIMARY KEY (`id`), 
	UNIQUE KEY `idx_file_hash` (`file_sha1`), 
	KEY `idx_status` (`status`)
) ENGINE = INNODB DEFAULT CHARSET = utf8
--mysql 
CREATE TABLE `res_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `email` varchar(255) DEFAULT NULL,
  `mobile` varchar(45) DEFAULT NULL,
  `nickname` varchar(45) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL,
  `head_image_url` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL COMMENT '自我描述',
  `name` varchar(45) DEFAULT NULL,
  `birth` date DEFAULT NULL,
  `province` varchar(45) DEFAULT NULL COMMENT '省',
  `city` varchar(45) DEFAULT NULL COMMENT '市',
  `company` varchar(45) DEFAULT NULL COMMENT '公司',
  `address` varchar(45) DEFAULT NULL COMMENT '地址',
  `sex` varchar(45) DEFAULT NULL COMMENT '性别',
  `contact_info` varchar(1000) DEFAULT NULL COMMENT '联系方式：如qq,msn,网站等 json方式保存{"key","value"}',
  `create_time` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='用户表';



-- postgresql
CREATE TABLE res_user
(
  id bigserial NOT NULL,
  email character varying(255),
  mobile character varying(45),
  nickname character varying(45),
  password character varying(255),
  head_image_url character varying(255),
  description character varying(255), -- 自我描述
  name character varying(45),
  birth timestamp with time zone,
  province character varying(45), -- 省
  city character varying(45), -- 市
  company character varying(45), -- 公司
  address character varying(45), -- 地址
  sex character varying(45), -- 性别
  contact_info character varying(1000), -- 联系方式：如qq,msn,网站等 json方式保存{"key","value"}
  create_time character varying(255),
  CONSTRAINT res_user_pkey PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
ALTER TABLE res_user
  OWNER TO postgres;
COMMENT ON TABLE res_user
  IS '用户表';
COMMENT ON COLUMN res_user.description IS '自我描述';
COMMENT ON COLUMN res_user.province IS '省';
COMMENT ON COLUMN res_user.city IS '市';
COMMENT ON COLUMN res_user.company IS '公司';
COMMENT ON COLUMN res_user.address IS '地址';
COMMENT ON COLUMN res_user.sex IS '性别';
COMMENT ON COLUMN res_user.contact_info IS '联系方式：如qq,msn,网站等 json方式保存{"key","value"}';





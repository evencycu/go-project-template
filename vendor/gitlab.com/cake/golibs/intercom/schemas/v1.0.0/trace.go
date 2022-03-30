package semconv

import "go.opentelemetry.io/otel/attribute"

const (
	M800CIDKey = attribute.Key("m800.cid")

	M800EIDKey = attribute.Key("m800.eid")

	M800NamespaceKey = attribute.Key("m800.namespace")

	M800ServiceKey     = attribute.Key("m800.service")
	M800ServiceHomeKey = attribute.Key("m800.service.home")

	M800ClientPlatformKey = attribute.Key("m800.client.platform")
	M800ClientDeviceIDKey = attribute.Key("m800.client.device.id")
	M800ClientWebIDKey    = attribute.Key("m800.client.web.id")

	M800UserRoleKey       = attribute.Key("m800.user.role")
	M800UserGroupKey      = attribute.Key("m800.user.group")
	M800UserAnonymousKey  = attribute.Key("m800.user.anonymous")
	M800UserCustomRoleKey = attribute.Key("m800.user.custom_role")

	// intercom httpDo
	M800ErrorCodeKey    = attribute.Key("m800.error")
	M800ErrorMessageKey = attribute.Key("m800.error.message")

	M800CrossFromNamespaceKey = attribute.Key("m800.cross.from.namespace")
	M800CrossToNamespaceKey   = attribute.Key("m800.cross.to.namespace")

	M800TagIDKey             = attribute.Key("m800.tag.id")
	M800RuleIDKey            = attribute.Key("m800.rule.id")
	M800StaffIDKey           = attribute.Key("m800.staff.id")
	M800ServiceIDKey         = attribute.Key("m800.service.id")
	M800InquiryIDKey         = attribute.Key("m800.inquiry.id")
	M800DestinationIDKey     = attribute.Key("m800.destination.id")
	M800AccessNumberIDKey    = attribute.Key("m800.access.number.id")
	M800ChannelIDKey         = attribute.Key("m800.channel.id")
	M800ChannelSourceIDKey   = attribute.Key("m800.channel.source.id")
	M800AnnouncementIDKey    = attribute.Key("m800.announcement.id")
	M800AccessAllowRuleIDKey = attribute.Key("m800.access.allow.rule.id")
	M800AccessBlockRuleIDKey = attribute.Key("m800.access.block.rule.id")
	M800CLIAllowRuleIDKey    = attribute.Key("m800.cli.allow.rule.id")
	M800SMSNumberIDKey       = attribute.Key("m800.sms.number.id")
)

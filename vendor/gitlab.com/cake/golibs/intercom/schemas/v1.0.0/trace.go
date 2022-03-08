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
)

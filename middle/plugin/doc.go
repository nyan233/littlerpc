package plugin

type InjectPlugin func(name string, plugin Plugin)

type InjectServerPlugin func(name string, plugin ServerPlugin)

type InjectClientPlugin func(name string, plugin ClientPlugin)

package migration

type MigrationPorts interface {
	Migrate()
	Seed()
}

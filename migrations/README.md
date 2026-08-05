# Migrations

Migration files live here. Create a new one with:

```bash
make migrate-create name=add_users_table
```

This creates `migrations/YYYYMMDDHHMMSS_add_users_table.sql` with Up and Down sections.

Apply pending migrations:

```bash
make migrate-up
```

The DSN is read automatically from `DB_DSN` in your environment / `.env` file.

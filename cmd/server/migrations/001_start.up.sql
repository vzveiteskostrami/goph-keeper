CREATE TABLE IF NOT EXISTS 
      UAUTH(
            USERID bigint not null,
            USER_NAME character varying(64) NOT NULL,
            USER_PWD character varying(64) NOT NULL,
            DELETE_FLAG boolean DEFAULT false
           );
CREATE UNIQUE INDEX IF NOT EXISTS uauth_unique_on_userid ON uauth (USERID);
CREATE INDEX IF NOT EXISTS uauth_on_user_name ON uauth (USER_NAME);

CREATE TABLE IF NOT EXISTS 
      DATUM(
             OID bigint not null,
             USERID bigint not null,
             DATA_TYPE smallint not null,
             DATA_NAME character varying(256),
             CREATE_TIME timestamptz not null, 
             UPDATE_TIME timestamptz not null, 
             DATA bytea,
             DELETE_FLAG boolean DEFAULT false
            );
CREATE UNIQUE INDEX IF NOT EXISTS datum_unique_on_oid ON datum (OID);
CREATE INDEX IF NOT EXISTS datum_on_userid ON datum (USERID);

create sequence if not exists gen_oid as bigint minvalue 1 no maxvalue start 1 no cycle;

COMMIT;

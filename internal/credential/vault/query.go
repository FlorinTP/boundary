package vault

const (
	insertTokenQuery = `
insert into credential_vault_token (
  token_hmac, -- $1
  token, -- $2
  store_id, -- $3
  key_id, -- $4
  status, -- $5
  last_renewal_time, -- $6
  expiration_time -- $7
) values (
  $1, -- token_hmac
  $2, -- token
  $3, -- store_id
  $4, -- key_id
  $5, -- status
  $6, -- last_renewal_time
  wt_add_seconds_to_now($7)  -- expiration_time
);
`

	insertCredentialWithExpirationQuery = `
insert into credential_vault_credential (
  public_id, -- $1
  library_id, -- $2
  session_id, -- $3
  token_hmac, -- $4
  external_id, -- $5
  is_renewable, -- $6
  last_renewal_time, -- $7
  expiration_time -- $8
) values (
  $1, -- public_id
  $2, -- library_id
  $3, -- session_id
  $4, -- token_hmac
  $5, -- external_id
  $6, -- is_renewable
  $7, -- last_renewal_time
  wt_add_seconds_to_now($8)  -- expiration_time
);
`

	insertCredentialWithInfiniteExpirationQuery = `
insert into credential_vault_credential (
  public_id, -- $1
  library_id, -- $2
  session_id, -- $3
  token_hmac, -- $4
  external_id, -- $5
  is_renewable, -- $6
  last_renewal_time, -- $7
  expiration_time -- infinity
) values (
  $1, -- public_id
  $2, -- library_id
  $3, -- session_id
  $4, -- token_hmac
  $5, -- external_id
  $6, -- is_renewable
  $7, -- last_renewal_time
  'infinity' -- expiration_time
);
`

	upsertClientCertQuery = `
insert into credential_vault_client_certificate
  (store_id, certificate, certificate_key, certificate_key_hmac, key_id)
values
  ($1, $2, $3, $4, $5)
on conflict (store_id) do update
  set certificate     = excluded.certificate,
      certificate_key = excluded.certificate_key,
      certificate_key_hmac = excluded.certificate_key_hmac,
      key_id          = excluded.key_id
returning *;
`

	deleteClientCertQuery = `
delete from credential_vault_client_certificate
 where store_id = $1;
`

	selectPrivateLibrariesQuery = `
select *
  from credential_vault_library_private
 where public_id in (%s);
`

	updateSessionCredentialQuery = `
update session_credential_dynamic
   set credential_id = $1
 where library_id = $2
   and session_id = $3
   and credential_purpose = $4
   and credential_id is null
returning *;
`

	updateTokenExpirationQuery = `
update credential_vault_token 
  set last_renewal_time = now(),
  expiration_time = wt_add_seconds_to_now(?)
  where token_hmac = ?;
`

	updateTokenStatusQuery = `
update credential_vault_token 
	set status = ? 
	where token_hmac = ?;
`

	tokenRenewalNextRunInQuery = `
select
	extract(epoch from (last_renewal_time + (expiration_time - last_renewal_time) / 2) - now())::int as renewal_in
  	from credential_vault_token
 	where expiration_time = (
	  select min(expiration_time)
  	    from credential_vault_token
	    where status in ('current', 'maintaining')
	);
`
)

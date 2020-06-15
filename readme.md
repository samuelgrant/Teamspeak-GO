# Boom Alliance TS3 API
A simple TS3 Server Query Client.

## Planned Commands
The following functions are planned

### User
* TS3Client.User.List()
* TS3Client.User.Find(uid)
* TS3Client.User.Search(name)
* TS3Client.User.CustomSearch(field, val)
* TS3Client.User.AssignServerGroup(uid, roleid)
* TS3Client.User.RemoveServerGroup(uid, roleid)
* TS3Client.User.Poke(uid, msg)
* TS3Client.User.Ban(uid, duration, reason, type) - Duration in seconds | Type IP || TS3 Client ID
* TS3Client.User.RevokeBan(uid)

### Permissions
* TS3Client.ChannelGroups - List all channel groups
* TS3Client.ChannelGroupsPoke(cid, cgid, msg) - poke the members of a specific channel group

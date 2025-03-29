# nt_gui




## License

This project is licensed under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0) with the [Commons Clause](https://commonsclause.com/), restricting commercial use without written permission from the author.


## DB details


** dns **
type: (not in db)
* seq
* Status
DNS_resolver (not in db, can find it in the nt cli in history)
DNS_query (not in db, can find it in the nt cli in history)
* DNS_response
* Record
DNS_protocol (not in db, can find it in the nt cli in history)
* Response_time(ms),
* SendDate,
* SendTime,
Packet Sent(sequence + 1 )
* successsponse,
* failure late,
* MinRTT
* Max RTT
* Avg RTT
* Additional Info

** http **
Type: (not in db)
* seq
* status
Method (not in db, get from ntcli)
url (not in db, get from ntcli)
* Response_code
* response_phase
* response_time(ms)
* sendDate
* sendTime
SessionSent (seq +1)
* sessionsuccess,
* failure rate
* minRTT
* MaxRTT
* AvgRTT
* additinalInfo

** tcp ** 
Type: (not in db)
* seq
* Status
DestHost (get from ntcli)
DestAddr (get from ntcli)
DestPort (get from ntcli)
payloadsize (get from ntcli)
* RTT (ms)
* SendDate
* SendTime
PacketSent (seq + 1)
* PacketRecv
* Packet Loss
* MinRTT
* AvgRTT
* MaxRTT
* AdditionalInfo

** icmp **
Type (not in db)
* seq
* status
DestHost (get from ntcli)
DesetAddr (get from ntcli)
Payloadsize (get from ntcli)
* RTT (ms)
* SendDate
* SendTime
PacketsSent (seq + 1 )
* Packet Recv
* Packet Loss
* MinRTT
* AvgRTT
* MaxRTT
* AdditionalInfo





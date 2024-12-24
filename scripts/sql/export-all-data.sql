SELECT ote.type, ot.ticket_id, otc.subject, ote.id ordem, ote.poster, ote.body from ost_ticket ot
inner join ost_thread ot2 on ot2.object_id = ot.ticket_id
INNER join ost_thread_entry ote on ote.thread_id = ot2.id
INNER JOIN ost_ticket__cdata otc on otc.ticket_id = ot.ticket_id
ORDER BY ot.ticket_id

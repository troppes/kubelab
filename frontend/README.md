Ziele: Frontend soll innerhalb des Clusters laufen, also SA nutzen

Erstmal nur User Flow:

User soll sich mit token einloggen, token hat die ID und anhand derer alles aus seinen NS ziehen


Danach

isTeacher hat zugriff auf alle deployments mit dem label der klasse: (RBAC geht da nicht, also listen am besten SSR)


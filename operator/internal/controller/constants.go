package controller

const storageClass = "kubelab-client"
const groupPrefix = "keycloak:"
const kubelabPrefix = "kubelab:"

// kubelabuser-controller constants
const userFinalizer = "kubeuser.kubelab.local/finalizer"
const roleBindingName = "user-rolebinding"
const claimNameUser = "user-claim"
const roleName = "user-role"

// classroom-controller constants
const classroomFinalizer = "classroom.kubelab.local/finalizer"
const classroomOwnerKey = ".metadata.namespace"
const userOwnerKey = ".spec.id"
const claimNameClass = "class-claim"

const nfsServer = "1.2.3.4"

import Swal from 'sweetalert2'
import type { Ref } from 'vue'

type Submit=(action:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>,confirmation?:string|false)=>Promise<boolean>

export function useUsers(options:{api:any,data:any,currentUser:Ref<any>,submit:Submit}){
  const {api,data,currentUser,submit}=options
  const userForm=reactive<any>({name:'',email:'',password:'',role:'cashier',active:true})
  const userModal=ref(false)
  const editingUser=ref<number|null>(null)

  async function loadUsers(){if(currentUser.value?.role==='admin')data.users=await api.users()}
  function openUser(user:any=null){editingUser.value=user?.id||null;Object.assign(userForm,user?{name:user.name,email:user.email,password:'',role:user.role,active:user.active}:{name:'',email:'',password:'',role:'cashier',active:true});userModal.value=true}
  function closeUser(){editingUser.value=null;userModal.value=false;Object.assign(userForm,{name:'',email:'',password:'',role:'cashier',active:true})}
  async function saveUser(){const id=editingUser.value;const body={...userForm};if(!body.password)delete body.password;if(await submit(()=>id?api.updateUser(id,body):api.createUser(body),`Pengguna berhasil ${id?'diubah':'ditambahkan'}`,[loadUsers]))closeUser()}
  async function removeUser(user:any){if(String(user.id)===String(currentUser.value.id)){await Swal.fire({icon:'warning',title:'Tidak dapat menghapus akun sendiri',confirmButtonColor:'#1d6b43'});return}const result=await Swal.fire({icon:'warning',title:`Hapus ${user.name}?`,text:'Pengguna tidak dapat login lagi.',showCancelButton:true,confirmButtonText:'Hapus',cancelButtonText:'Batal',confirmButtonColor:'#b8322a'});if(result.isConfirmed)await submit(()=>api.deleteUser(user.id),'Pengguna berhasil dihapus',[loadUsers],false)}

  return {userForm,userModal,editingUser,loadUsers,openUser,closeUser,saveUser,removeUser}
}

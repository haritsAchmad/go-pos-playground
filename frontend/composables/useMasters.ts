type Submit=(action:()=>Promise<any>,message:string,reload:Array<()=>Promise<void>>,confirmation?:string|false)=>Promise<boolean>

export function useMasters(options:{api:any,data:any,editing:any,submit:Submit,reloadItems:()=>Promise<void>}){
  const {api,data,editing,submit,reloadItems}=options
  const masterForm=reactive({table:'categories',name:''})

  async function loadMasters(){const [categories,brands,units,payment_methods]=await Promise.all([api.masters('categories'),api.masters('brands'),api.masters('units'),api.masters('payment_methods')]);Object.assign(data,{categories,brands,units,payment_methods})}
  function openMaster(table:string){masterForm.table=table;navigateTo('/pengaturan')}
  async function saveMaster(){const id=editing.master;if(await submit(()=>id?api.updateMaster(masterForm.table,id,masterForm):api.createMaster(masterForm.table,masterForm),`Data berhasil ${id?'diubah':'ditambahkan'}`,[loadMasters,reloadItems])){editing.master=null;masterForm.name=''}}
  function editMaster(table:string,v:any){masterForm.table=table;masterForm.name=v.name;editing.master=v.id}
  async function deleteMaster(table:string,v:any){if(confirm(`Hapus ${v.name}?`))await submit(()=>api.deleteMaster(table,v.id),'Data berhasil dihapus',[loadMasters,reloadItems],false)}
  function findByName(values:any[],name:string){return values.find((v:any)=>String(v.name).toLowerCase()===String(name).toLowerCase())?.id||null}
  async function findOrCreateMaster(table:'categories'|'brands'|'units',values:any[],name:string){
    if(!String(name||'').trim())return null
    const existing=findByName(values,name);if(existing)return existing
    await api.createMaster(table,{name:String(name).trim()})
    const refreshed=await api.masters(table);values.splice(0,values.length,...refreshed)
    return findByName(values,name)
  }

  return {masterForm,loadMasters,openMaster,saveMaster,editMaster,deleteMaster,findOrCreateMaster}
}

<script setup lang="ts">
import Swal from 'sweetalert2'

const api = useKoperasiApi()
const nav = [
  ['dashboard', 'Dashboard'], ['items', 'Barang'], ['customers', 'Pelanggan'], ['suppliers', 'Supplier'],
  ['sale', 'Kasir'], ['purchase', 'Pembelian'], ['history', 'Histori'], ['debts', 'Piutang'], ['masters', 'Pengaturan Barang'],
]
const active = ref('dashboard')
const dashboardYear = ref(new Date().getFullYear())
const years = Array.from({length:5},(_,i)=>new Date().getFullYear()-i)
const loading = ref(false)
const error = ref('')
const notice = ref('')
const data = reactive<any>({ dashboard:{},items:[],customers:[],suppliers:[],categories:[],brands:[],units:[],payment_methods:[],transactions:[],debts:[] })
const itemForm = reactive<any>({ sku:'',name:'',description:'',supplier_id:null,category_id:null,brand_id:null,unit_id:null,stock:0,cost:0,price:0 })
const customerForm = reactive<any>({ code:'',name:'',phone:'',address:'' })
const supplierForm = reactive<any>({ name:'',phone:'',address:'' })
const transactionForm = reactive<any>({ customer_id:null,supplier_id:null,payment_method_id:null,paid_amount:0,notes:'',items:[{item_id:null,quantity:1,unit_price:0}] })
const masterForm = reactive({ table:'categories', name:'' })
const editing = reactive<any>({ item:null, customer:null, supplier:null, master:null })
const modal = ref<null|'item'|'customer'|'supplier'>(null)
const filters = reactive({ itemSearch:'', category:'', stock:'', customerSearch:'', supplierSearch:'' })
const months = ['Jan','Feb','Mar','Apr','Mei','Jun','Jul','Agu','Sep','Okt','Nov','Des']
const maxSales = computed(()=>Math.max(...(data.dashboard.monthly_sales||[0]),1))

const money = (v:number) => new Intl.NumberFormat('id-ID',{style:'currency',currency:'IDR',maximumFractionDigits:0}).format(v||0)
const itemById = (id:number) => data.items.find((v:any)=>v.id===Number(id))
const lineTotal = computed(()=>transactionForm.items.reduce((sum:number,line:any)=>sum+(Number(line.quantity)||0)*(Number(line.unit_price)||0),0))
const changeAmount = computed(()=>Number(transactionForm.paid_amount||0)-lineTotal.value)
const filteredItems = computed(()=>data.items.filter((v:any)=>{
  const query=filters.itemSearch.toLowerCase();const matchesQuery=!query||[v.sku,v.name,v.category_name,v.brand_name].some(x=>String(x||'').toLowerCase().includes(query));
  const matchesCategory=!filters.category||String(v.category_id)===filters.category;
  const matchesStock=!filters.stock||(filters.stock==='empty'?Number(v.stock)===0:filters.stock==='low'?Number(v.stock)>0&&Number(v.stock)<=5:Number(v.stock)>5);
  return matchesQuery&&matchesCategory&&matchesStock
}))
const filteredCustomers = computed(()=>data.customers.filter((v:any)=>!filters.customerSearch||[v.code,v.name,v.phone].some(x=>String(x||'').toLowerCase().includes(filters.customerSearch.toLowerCase()))))
const filteredSuppliers = computed(()=>data.suppliers.filter((v:any)=>!filters.supplierSearch||[v.name,v.phone,v.address].some(x=>String(x||'').toLowerCase().includes(filters.supplierSearch.toLowerCase()))))

async function load() {
  loading.value=true; error.value=''
  try {
    const [dashboard,items,customers,suppliers,categories,brands,units,payment_methods,transactions,debts] = await Promise.all([
      api.dashboard(dashboardYear.value),api.items(),api.customers(),api.suppliers(),api.masters('categories'),api.masters('brands'),api.masters('units'),api.masters('payment_methods'),api.transactions(),api.debts(),
    ])
    Object.assign(data,{dashboard,items,customers,suppliers,categories,brands,units,payment_methods,transactions,debts})
    if (!transactionForm.customer_id) transactionForm.customer_id=customers.find((c:any)=>c.code==='UMUM')?.id
  } catch(e:any){error.value=e?.data?.message||e.message||'Gagal memuat data'} finally{loading.value=false}
}
async function submit(action:()=>Promise<any>, message:string, confirmation:string|false=false){if(confirmation!==false){const result=await Swal.fire({icon:'question',title:'Konfirmasi',text:confirmation,showCancelButton:true,confirmButtonText:'Ya, simpan',cancelButtonText:'Batal',confirmButtonColor:'#1d6b43'});if(!result.isConfirmed)return false}error.value='';notice.value='';try{await action();await load();await Swal.fire({icon:'success',title:'Berhasil',text:message,confirmButtonColor:'#1d6b43'});return true}catch(e:any){const message=e?.data?.message||e.message||'Gagal menyimpan data';await Swal.fire({icon:'error',title:'Gagal',text:message,confirmButtonColor:'#1d6b43'});return false}}
watch(active,()=>{error.value='';notice.value='';modal.value=null})
watch(dashboardYear,load)
watch(()=>customerForm.phone,v=>{const clean=String(v||'').replace(/\D/g,'');if(v!==clean)customerForm.phone=clean})
watch(()=>supplierForm.phone,v=>{const clean=String(v||'').replace(/\D/g,'');if(v!==clean)supplierForm.phone=clean})
function nullableNumber(v:any){return v?Number(v):null}
async function saveItem(){const payload={...itemForm,supplier_id:nullableNumber(itemForm.supplier_id),category_id:nullableNumber(itemForm.category_id),brand_id:nullableNumber(itemForm.brand_id),unit_id:nullableNumber(itemForm.unit_id),stock:Number(itemForm.stock),cost:Number(itemForm.cost),price:Number(itemForm.price)};if(await submit(()=>editing.item?api.updateItem(editing.item,payload):api.createItem(payload),editing.item?'Barang berhasil diubah':'Barang berhasil ditambahkan'))cancelItem()}
function openItem(v:any=null){cancelItem(false);if(v){editing.item=v.id;Object.assign(itemForm,v)}modal.value='item'}
function editItem(v:any){openItem(v)}
function cancelItem(close=true){editing.item=null;Object.assign(itemForm,{sku:'',name:'',description:'',supplier_id:null,category_id:null,brand_id:null,unit_id:null,stock:0,cost:0,price:0});if(close)modal.value=null}
async function removeItem(v:any){const result=await Swal.fire({icon:'warning',title:`Hapus ${v.name}?`,text:'Data yang sudah dihapus tidak dapat dikembalikan.',showCancelButton:true,confirmButtonText:'Hapus',cancelButtonText:'Batal',confirmButtonColor:'#b8322a'});if(result.isConfirmed)await submit(()=>api.deleteItem(v.id),'Barang berhasil dihapus')}
async function saveParty(kind:'customer'|'supplier'){const form=kind==='customer'?customerForm:supplierForm;const id=editing[kind];if(await submit(()=>kind==='customer'?(id?api.updateCustomer(id,form):api.createCustomer(form)):(id?api.updateSupplier(id,form):api.createSupplier(form)),`${kind==='customer'?'Pelanggan':'Supplier'} berhasil ${id?'diubah':'ditambahkan'}`))closeParty(kind)}
function openParty(kind:'customer'|'supplier',v:any=null){if(v?.code==='UMUM')return;editing[kind]=v?.id||null;const form=kind==='customer'?customerForm:supplierForm;Object.keys(form).forEach(k=>form[k]='');if(v)Object.assign(form,v);modal.value=kind}
function editParty(kind:'customer'|'supplier',v:any){openParty(kind,v)}
function closeParty(kind:'customer'|'supplier'){editing[kind]=null;const form=kind==='customer'?customerForm:supplierForm;Object.keys(form).forEach(k=>form[k]='');modal.value=null}
async function removeParty(kind:'customer'|'supplier',v:any){if(v.code==='UMUM')return;const result=await Swal.fire({icon:'warning',title:`Hapus ${v.name}?`,showCancelButton:true,confirmButtonText:'Hapus',cancelButtonText:'Batal',confirmButtonColor:'#b8322a'});if(result.isConfirmed)await submit(()=>kind==='customer'?api.deleteCustomer(v.id):api.deleteSupplier(v.id),`${kind==='customer'?'Pelanggan':'Supplier'} berhasil dihapus`)}
function openMaster(table:string){masterForm.table=table;active.value='masters'}
function blockInvalidNumber(event:KeyboardEvent){const target=event.target as HTMLInputElement;if(target?.type==='number'&&['e','E','+','-','.'].includes(event.key))event.preventDefault()}
function sanitizeNumeric(event:Event){const target=event.target as HTMLInputElement;if(target?.type==='tel')target.value=target.value.replace(/\D/g,'')}
async function voidTransaction(v:any){const reason=prompt('Alasan pembatalan transaksi (minimal 5 karakter):');if(reason===null)return;if(reason.trim().length<5){error.value='Alasan pembatalan minimal 5 karakter.';return};await submit(()=>api.voidTransaction(v.id,reason.trim()),'Transaksi dibatalkan dan stok sudah dikembalikan')}
async function saveMaster(){const id=editing.master;if(await submit(()=>id?api.updateMaster(masterForm.table,id,masterForm):api.createMaster(masterForm.table,masterForm),`Data berhasil ${id?'diubah':'ditambahkan'}`)){editing.master=null;masterForm.name=''}}
function editMaster(table:string,v:any){masterForm.table=table;masterForm.name=v.name;editing.master=v.id}
async function deleteMaster(table:string,v:any){if(confirm(`Hapus ${v.name}?`))await submit(()=>api.deleteMaster(table,v.id),'Data berhasil dihapus',false)}
function chooseItem(line:any){const item=itemById(line.item_id);if(item)line.unit_price=active.value==='purchase'?item.cost:item.price}
function addLine(){transactionForm.items.push({item_id:null,quantity:1,unit_price:0})}
function resetTransaction(){Object.assign(transactionForm,{customer_id:data.customers.find((c:any)=>c.code==='UMUM')?.id||null,supplier_id:null,payment_method_id:null,paid_amount:0,notes:'',items:[{item_id:null,quantity:1,unit_price:0}]})}
async function saveTransaction(type:'SALE'|'PURCHASE') { if(await submit(()=>api.createTransaction({...transactionForm,transaction_type:type,customer_id:type==='SALE'?Number(transactionForm.customer_id):null,supplier_id:type==='PURCHASE'?Number(transactionForm.supplier_id):null,payment_method_id:transactionForm.payment_method_id?Number(transactionForm.payment_method_id):null,items:transactionForm.items.map((v:any)=>({...v,item_id:Number(v.item_id),quantity:Number(v.quantity),unit_price:Number(v.unit_price)}))}),`${type==='SALE'?'Penjualan':'Pembelian'} berhasil dicatat`, `Simpan ${type==='SALE'?'penjualan':'pembelian'} sebesar ${money(lineTotal.value)}?`))resetTransaction() }
onMounted(load)
</script>

<template>
  <div class="app-shell">
    <aside><div class="brand"><span>K</span><div><strong>Koperasi</strong><small>Operational Console</small></div></div><nav><button v-for="n in nav" :key="n[0]" :class="{active:active===n[0]}" @click="active=n[0]">{{ n[1] }}</button></nav></aside>
    <main @keydown="blockInvalidNumber" @input="sanitizeNumeric">
      <header><div><p class="eyebrow">SIG KOPERASI · GO + NUXT</p><h1>{{ nav.find(n=>n[0]===active)?.[1] }}</h1></div><button class="soft" @click="load">{{ loading?'Memuat…':'Refresh' }}</button></header>
      <p v-if="error" class="alert error">{{ error }}</p><p v-if="notice" class="alert success">{{ notice }}</p>

      <template v-if="active==='dashboard'">
        <section class="stats"><article><small>Penjualan Hari Ini</small><strong>{{ money(data.dashboard.today_sales) }}</strong></article><article><small>Pembelian Hari Ini</small><strong>{{ money(data.dashboard.today_purchases) }}</strong></article><article><small>Piutang Terbuka</small><strong>{{ money(data.dashboard.open_debt) }}</strong></article><article><small>Stok Menipis</small><strong>{{ data.dashboard.low_stock_items||0 }} barang</strong></article><article><small>Total Barang</small><strong>{{ data.dashboard.total_items||0 }}</strong></article><article><small>Pelanggan</small><strong>{{ data.dashboard.total_customers||0 }}</strong></article></section>
        <section class="panel chart-panel"><div class="section-title"><div><h2>Grafik Penjualan Tahunan</h2><small>Total nilai transaksi penjualan per bulan</small></div><select v-model.number="dashboardYear"><option v-for="year in years" :value="year">{{year}}</option></select></div><div class="chart"><div v-for="(value,i) in data.dashboard.monthly_sales||[]" class="bar-column"><strong>{{value?money(value):'Rp0'}}</strong><div class="bar-track"><span :style="{height:`${Math.max((value/maxSales)*100,value?4:0)}%`}"></span></div><small>{{months[i]}}</small></div></div></section>
      </template>

      <template v-else-if="active==='items'">
        <div class="page-heading"><div><h2>Barang</h2><p>Kelola katalog dan persediaan barang.</p></div><button class="primary" @click="openItem()">+ Tambah Barang</button></div>
        <section class="panel toolbar"><input v-model="filters.itemSearch" type="search" placeholder="Cari nama atau SKU..."><select v-model="filters.category"><option value="">Semua kategori</option><option v-for="v in data.categories" :value="String(v.id)">{{v.name}}</option></select><select v-model="filters.stock"><option value="">Semua stok</option><option value="empty">Stok habis</option><option value="low">Stok menipis (1-5)</option><option value="ready">Stok tersedia</option></select><button class="soft" @click="Object.assign(filters,{itemSearch:'',category:'',stock:''})">Reset Filter</button></section>
        <DataTable actions :rows="filteredItems" :columns="[['sku','Kode / SKU'],['name','Barang'],['category_name','Kategori'],['unit_name','Satuan'],['stock','Stok'],['cost','Harga Beli','money'],['price','Harga Jual','money']]" @edit="editItem" @delete="removeItem" />
      </template>

      <template v-else-if="active==='customers' || active==='suppliers'">
        <div class="page-heading"><div><h2>{{active==='customers'?'Pelanggan':'Supplier'}}</h2><p>Kelola data {{active==='customers'?'pelanggan dan anggota':'mitra pemasok'}}.</p></div><button class="primary" @click="openParty(active==='customers'?'customer':'supplier')">+ Tambah {{active==='customers'?'Pelanggan':'Supplier'}}</button></div>
        <section class="panel toolbar party-toolbar"><input v-if="active==='customers'" v-model="filters.customerSearch" type="search" placeholder="Cari kode, nama, atau telepon..."><input v-else v-model="filters.supplierSearch" type="search" placeholder="Cari nama, telepon, atau alamat..."></section>
        <DataTable v-if="active==='customers'" actions :rows="filteredCustomers" :columns="[['code','Kode'],['name','Nama'],['phone','Telepon'],['address','Alamat']]" @edit="editParty('customer',$event)" @delete="removeParty('customer',$event)"/><DataTable v-else actions :rows="filteredSuppliers" :columns="[['name','Nama'],['phone','Telepon'],['address','Alamat']]" @edit="editParty('supplier',$event)" @delete="removeParty('supplier',$event)"/>
      </template>

      <template v-else-if="active==='sale'||active==='purchase'">
        <section class="panel transaction"><h2>{{active==='sale'?'Transaksi Kasir':'Penerimaan Pembelian'}}</h2><p class="hint">{{active==='sale'?'Jika pembeli bukan anggota, gunakan pelanggan UMUM. Bila jumlah dibayar kurang dari total, sisanya otomatis menjadi piutang.':'Harga beli pada transaksi ini disimpan sebagai riwayat dan tidak berubah saat harga master barang diperbarui.'}}</p><div class="grid"><label v-if="active==='sale'">Pelanggan<select v-model="transactionForm.customer_id" required><option v-for="v in data.customers" :value="v.id">{{v.code}} · {{v.name}}</option></select></label><label v-else>Supplier<select v-model="transactionForm.supplier_id" required><option :value="null" disabled>Pilih supplier</option><option v-for="v in data.suppliers" :value="v.id">{{v.name}}</option></select></label><label>Metode Pembayaran<select v-model="transactionForm.payment_method_id" required><option :value="null" disabled>Pilih metode pembayaran</option><option v-for="v in data.payment_methods" :value="v.id">{{v.name}}</option></select></label><label>Jumlah Dibayar (Rp)<input v-model.number="transactionForm.paid_amount" type="number" min="0" :max="lineTotal" placeholder="Masukkan nominal pembayaran" required></label></div><div class="line-head"><span>Barang</span><span>Jumlah</span><span>Harga Satuan (Rp)</span><span>Subtotal</span><span></span></div><div class="lines"><div v-for="(line,i) in transactionForm.items" :key="i" class="line"><select v-model="line.item_id" required @change="chooseItem(line)"><option :value="null" disabled>Pilih barang</option><option v-for="v in data.items" :value="v.id">{{v.sku}} · {{v.name}} (stok {{v.stock}})</option></select><input v-model.number="line.quantity" aria-label="Jumlah barang" type="number" min="1" placeholder="Jumlah" required><input v-model.number="line.unit_price" aria-label="Harga satuan" type="number" min="0" placeholder="Harga satuan" required><strong>{{money(line.quantity*line.unit_price)}}</strong><button type="button" class="icon" :disabled="transactionForm.items.length===1" @click="transactionForm.items.splice(i,1)">×</button></div></div><button type="button" class="soft" @click="addLine">+ Tambah Barang</button><label>Catatan<textarea v-model="transactionForm.notes" placeholder="Opsional: catatan transaksi"/></label><div class="checkout"><div><small>Grand Total</small><strong>{{money(lineTotal)}}</strong></div><button class="primary" :disabled="!lineTotal" @click="saveTransaction(active==='sale'?'SALE':'PURCHASE')">Simpan Transaksi</button></div></section>
      </template>

      <DataTable v-else-if="active==='history'" void-actions :rows="data.transactions" :columns="[['invoice_no','Invoice'],['transaction_date','Tanggal','date'],['transaction_type','Tipe'],['customer_name','Pelanggan'],['supplier_name','Supplier'],['status','Status'],['grand_total','Total','money'],['amount_received','Uang Diterima','money'],['change_amount','Kembalian','money']]" @void="voidTransaction"/>
      <section v-else-if="active==='debts'" class="panel"><h2>Pembayaran Piutang Pelanggan</h2><p class="hint">Piutang dibuat otomatis ketika jumlah yang dibayar di kasir lebih kecil dari grand total. Masukkan nominal cicilan pada transaksi yang sesuai; sisa piutang akan berkurang otomatis.</p><div class="debt" v-for="v in data.debts" :key="v.id"><div><strong>{{v.customer_name}}</strong><small>Invoice {{v.invoice_no}}</small><small>Awal {{money(v.original_amount)}} · Sisa <b>{{money(v.remaining_amount)}}</b></small></div><form v-if="v.status==='OPEN'" @submit.prevent="submit(()=>api.payDebt(v.id,{amount:Number(($event.target as HTMLFormElement).amount.value)}),'Pembayaran piutang berhasil dicatat')"><label>Nominal Pembayaran (Rp)<input name="amount" type="number" min="1" :max="v.remaining_amount" placeholder="Maksimal sebesar sisa piutang" required></label><button class="primary">Catat Pembayaran</button></form><span v-else class="paid">Lunas</span></div><p v-if="!data.debts.length" class="empty">Belum ada piutang pelanggan.</p></section>
      <section v-else-if="active==='masters'" class="panel"><h2>Pengaturan Barang</h2><p class="hint">Kelola pilihan yang muncul pada form barang dan transaksi: kategori untuk mengelompokkan barang, merek untuk produsennya, satuan untuk bentuk stok, dan metode pembayaran untuk kasir.</p><form class="inline" @submit.prevent="saveMaster"><select v-model="masterForm.table" @change="editing.master=null;masterForm.name=''"><option value="categories">Kategori</option><option value="brands">Merek</option><option value="units">Satuan</option><option value="payment_methods">Metode Pembayaran</option></select><input v-model.trim="masterForm.name" minlength="2" maxlength="100" placeholder="Masukkan nama" required><button class="primary">{{editing.master?'Simpan':'Tambah'}}</button></form><div class="master-grid"><article v-for="table in ['categories','brands','units','payment_methods']"><h3>{{({categories:'Kategori',brands:'Merek',units:'Satuan',payment_methods:'Metode Pembayaran'} as any)[table]}}</h3><div class="master-row" v-for="v in data[table]"><span>{{v.name}}</span><button class="mini" @click="editMaster(table,v)">Ubah</button><button class="mini danger" @click="deleteMaster(table,v)">Hapus</button></div></article></div></section>
      <section v-if="active==='sale'||active==='purchase'" class="panel payment-summary"><div><small>Total Transaksi</small><strong>{{money(lineTotal)}}</strong></div><div><small>Uang Diterima</small><strong>{{money(transactionForm.paid_amount)}}</strong></div><div :class="changeAmount<0?'shortage':'change'"><small>{{changeAmount<0?'Kurang Bayar':'Kembalian'}}</small><strong>{{money(changeAmount)}}</strong></div></section>

      <div v-if="modal" class="modal-backdrop" @mousedown.self="modal==='item'?cancelItem():closeParty(modal)"><section class="modal-card" role="dialog" aria-modal="true"><div class="modal-head"><div><h2>{{editing[modal]?'Ubah':'Tambah'}} {{modal==='item'?'Barang':modal==='customer'?'Pelanggan':'Supplier'}}</h2><p>Lengkapi data berikut, lalu simpan perubahan.</p></div><button class="modal-close" aria-label="Tutup" @click="modal==='item'?cancelItem():closeParty(modal)">×</button></div>
        <form v-if="modal==='item'" class="grid" @submit.prevent="saveItem"><label>Kode / SKU<input v-model.trim="itemForm.sku" pattern="[A-Za-z0-9._-]{2,50}" placeholder="Contoh: ATK-001" required></label><label>Nama<input v-model.trim="itemForm.name" minlength="3" maxlength="100" required></label><label>Kategori<select v-model="itemForm.category_id" required><option :value="null" disabled>Pilih kategori</option><option v-for="v in data.categories" :value="v.id">{{v.name}}</option></select></label><label>Merek<select v-model="itemForm.brand_id" required><option :value="null" disabled>Pilih merek</option><option v-for="v in data.brands" :value="v.id">{{v.name}}</option></select></label><label>Satuan<select v-model="itemForm.unit_id" required><option :value="null" disabled>Pilih satuan</option><option v-for="v in data.units" :value="v.id">{{v.name}}</option></select></label><label>Supplier<select v-model="itemForm.supplier_id"><option :value="null">Tanpa supplier</option><option v-for="v in data.suppliers" :value="v.id">{{v.name}}</option></select></label><label>Stok awal<input v-model.number="itemForm.stock" type="number" min="0" required></label><label>Harga beli<input v-model.number="itemForm.cost" type="number" min="0" required></label><label>Harga jual<input v-model.number="itemForm.price" type="number" min="0" required></label><label class="wide">Deskripsi<textarea v-model="itemForm.description" maxlength="500"/></label><div class="modal-actions wide"><button type="button" class="soft" @click="cancelItem">Batal</button><button class="primary">{{editing.item?'Simpan Perubahan':'Simpan Barang'}}</button></div></form>
        <form v-else-if="modal==='customer'" class="grid" @submit.prevent="saveParty('customer')"><label>Kode Anggota<input v-model.trim="customerForm.code" pattern="[A-Za-z0-9._-]{2,50}" placeholder="Contoh: AGT-001" required></label><label>Nama<input v-model.trim="customerForm.name" minlength="3" maxlength="120" required></label><label>Nomor Telepon<input v-model="customerForm.phone" type="tel" inputmode="numeric" pattern="[0-9]{8,20}" placeholder="Contoh: 081234567890"></label><label class="wide">Alamat<textarea v-model="customerForm.address" maxlength="500"/></label><div class="modal-actions wide"><button type="button" class="soft" @click="closeParty('customer')">Batal</button><button class="primary">{{editing.customer?'Simpan Perubahan':'Simpan Pelanggan'}}</button></div></form>
        <form v-else class="grid" @submit.prevent="saveParty('supplier')"><label>Nama<input v-model.trim="supplierForm.name" minlength="3" maxlength="100" required></label><label>Nomor Telepon<input v-model="supplierForm.phone" type="tel" inputmode="numeric" pattern="[0-9]{8,20}" placeholder="Contoh: 081234567890" required></label><label class="wide">Alamat<textarea v-model="supplierForm.address" minlength="5" required/></label><div class="modal-actions wide"><button type="button" class="soft" @click="closeParty('supplier')">Batal</button><button class="primary">{{editing.supplier?'Simpan Perubahan':'Simpan Supplier'}}</button></div></form>
      </section></div>
    </main>
  </div>
</template>

export function useDashboard(api:any,data:any){
  const dashboardYear=ref(new Date().getFullYear())
  const dashboardMonth=ref(new Date().getMonth()+1)
  const years=Array.from({length:5},(_,i)=>new Date().getFullYear()-i)
  const months=['Jan','Feb','Mar','Apr','Mei','Jun','Jul','Agu','Sep','Okt','Nov','Des']
  const maxSales=computed(()=>Math.max(...(data.dashboard.monthly_sales||[0]),1))
  const jakartaYearMonth=(value:string)=>{const parts=new Intl.DateTimeFormat('en-CA',{year:'numeric',month:'numeric',timeZone:'Asia/Jakarta'}).formatToParts(new Date(value));return {year:Number(parts.find(v=>v.type==='year')?.value),month:Number(parts.find(v=>v.type==='month')?.value)}}
  const monthlyReportTransactions=computed(()=>data.transactions.filter((v:any)=>{const date=jakartaYearMonth(v.transaction_date);return v.transaction_type==='SALE'&&v.status==='ACTIVE'&&date.year===dashboardYear.value&&date.month===dashboardMonth.value}))
  const monthlyReportTotal=computed(()=>monthlyReportTransactions.value.reduce((sum:number,v:any)=>sum+Number(v.grand_total||0),0))
  const dashboardPeriods=computed(()=>[
    {label:'Hari Ini',value:data.dashboard.today||{}},
    {label:'Kemarin',value:data.dashboard.yesterday||{}},
    {label:`${months[dashboardMonth.value-1]} ${dashboardYear.value}`,value:data.dashboard.selected_month||{}},
    {label:`Tahun ${dashboardYear.value}`,value:data.dashboard.selected_year||{}},
  ])
  const dailyTotals=computed(()=>(data.dashboard.daily||[]).reduce((sum:any,row:any)=>({income:sum.income+Number(row.income||0),expense:sum.expense+Number(row.expense||0),debt:sum.debt+Number(row.debt||0),net_income:sum.net_income+Number(row.net_income||0)}),{income:0,expense:0,debt:0,net_income:0}))

  async function loadDashboard(){data.dashboard=await api.dashboard(dashboardYear.value,dashboardMonth.value)}

  return {dashboardYear,dashboardMonth,years,months,maxSales,monthlyReportTransactions,monthlyReportTotal,dashboardPeriods,dailyTotals,loadDashboard}
}

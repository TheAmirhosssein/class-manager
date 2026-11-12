/* Shared API boundary. UI never calls fetch directly. Response contracts: docs/API.md. */
(() => {
  'use strict';
  const C = window.APP_CONFIG;
  const seed = {
    school: { name: 'مدرسه نمونه' },
    courses: [
      { id:'math10', grade:'10', title:'ریاضی دهم؛ از پایه تا تسلط', subject:'ریاضی ۱', description:'از مجموعه‌ها تا تابع؛ مفهومی و قدم‌به‌قدم.', lessons:24, duration:'۱۸ ساعت', type:'آموزش + تمرین', topics:['مجموعه، الگو و دنباله','مثلثات و توان‌های گویا','معادله، نامعادله و تابع'] },
      { id:'math11', grade:'11', title:'حسابان یازدهم؛ عمیق‌تر یاد بگیر', subject:'حسابان ۱', description:'تابع‌ها را بشناس، مسئله‌ها را با اطمینان حل کن.', lessons:28, duration:'۲۱ ساعت', type:'آموزش + تمرین', topics:['جبر و معادله','تابع، توابع نمایی و لگاریتمی','مثلثات، حد و پیوستگی'] },
      { id:'math12', grade:'12', title:'حسابان دوازدهم؛ یک قدم جلوتر', subject:'حسابان ۲', description:'از حد تا مشتق؛ آماده برای چالش‌های بزرگ‌تر.', lessons:32, duration:'۲۴ ساعت', type:'آموزش + تمرین', topics:['تابع و مثلثات','حدهای نامتناهی و حد در بی‌نهایت','مشتق و کاربردهای آن'] }
    ],
    classes: [ {id:'c10',title:'ریاضی دهم — الف',grade:'10',room:'کلاس ۱۰۱',studentCount:6}, {id:'c11',title:'حسابان یازدهم — الف',grade:'11',room:'کلاس ۲۰۲',studentCount:4}, {id:'c12',title:'حسابان دوازدهم — الف',grade:'12',room:'کلاس ۳۰۱',studentCount:4} ],
    schedule: [ {id:'t1',classId:'c10',day:0,start:'08:00',end:'09:30',topic:'مجموعه‌ها و الگوها'}, {id:'t2',classId:'c11',day:2,start:'10:00',end:'11:30',topic:'آشنایی با تابع'}, {id:'t3',classId:'c12',day:4,start:'08:00',end:'09:30',topic:'حد و پیوستگی'} ],
    students: {c10:[{id:'s1',name:'آراد احمدی'},{id:'s2',name:'علی رضایی'},{id:'s3',name:'پارسا محمدی'},{id:'s4',name:'امیر حسینی'},{id:'s5',name:'سامان کریمی'},{id:'s6',name:'محمد موسوی'}],c11:[{id:'s7',name:'مهدی نوری'},{id:'s8',name:'رضا اکبری'},{id:'s9',name:'اشکان رحیمی'},{id:'s10',name:'حسین امینی'}],c12:[{id:'s11',name:'نیما مرادی'},{id:'s12',name:'کیان زمانی'},{id:'s13',name:'ایلیا صادقی'},{id:'s14',name:'علی کاظمی'}]},
    attendance:{},discipline:[]
  };
  let db;
  try {
    const saved=JSON.parse(localStorage.getItem('classino.demo.v1')||localStorage.getItem('riazino.demo.v1'));
    db=saved&&saved.version===1?saved.data:structuredClone(seed);
  } catch { db=structuredClone(seed); }
  const copy=x=>structuredClone(x);
  const persist=()=>{try{localStorage.setItem('classino.demo.v1',JSON.stringify({version:1,data:db}));}catch{throw new Error('ذخیره‌سازی مرورگر در دسترس نیست. تنظیمات مرورگر را بررسی کنید.');}};
  const TOKEN_KEY='classino.accessToken';
  const token=()=>{try{return sessionStorage.getItem(TOKEN_KEY)||'';}catch{return '';}};
  const setToken=t=>{try{if(t)sessionStorage.setItem(TOKEN_KEY,t);else sessionStorage.removeItem(TOKEN_KEY);}catch{}};
  async function request(path,{method='GET',body}={}) {
    const headers={Accept:'application/json'};
    const access=token();
    if(access)headers.Authorization='Bearer '+access;
    if(body!==undefined)headers['Content-Type']='application/json';
    if(method!=='GET') {
      const item=document.cookie.split('; ').find(x=>x.startsWith(C.csrfCookie+'='));
      if(item)headers[C.csrfHeader]=decodeURIComponent(item.slice(C.csrfCookie.length+1));
    }
    const controller=new AbortController(),timeout=setTimeout(()=>controller.abort(),15000);
    try {
      const response=await fetch(C.apiBase.replace(/\/$/,'')+path,{method,headers,credentials:C.credentials,body:body===undefined?undefined:JSON.stringify(body),signal:controller.signal});
      const data=response.status===204?null:await response.json().catch(()=>null);
      if(!response.ok)throw new Error(data?.message || (response.status===401?'برای ادامه وارد حساب خود شوید.':response.status===403?'اجازه دسترسی به این بخش را ندارید.':'دریافت یا ذخیره اطلاعات انجام نشد. دوباره تلاش کنید.'));
      if(response.status!==204&&data===null)throw new Error('پاسخ سرور باید JSON معتبر باشد.');
      return data;
    } catch(e) { if(e.name==='AbortError')throw new Error('زمان پاسخ‌گویی سرور تمام شد. دوباره تلاش کنید.');if(e instanceof TypeError)throw new Error('ارتباط با سرور برقرار نشد. اتصال و آدرس API را بررسی کنید.');throw e; }
    finally {clearTimeout(timeout);}
  }
  const choose=(demo,live)=>C.demo?Promise.resolve().then(demo):live();
  const enc=encodeURIComponent;
  window.API={
    courses:()=>choose(()=>copy(db.courses),()=>request('/courses')),
    classes:()=>choose(()=>copy(db.classes),()=>request('/classes')),
    school:()=>choose(()=>copy(db.school),()=>request('/school')),
    saveSchool:body=>choose(()=>{db.school=copy(body);persist();return copy(body);},()=>request('/school',{method:'PATCH',body})),
    schedule:()=>choose(()=>copy(db.schedule),()=>request('/schedule')),
    addSchedule:body=>choose(()=>{const item={...copy(body),id:crypto.randomUUID()};db.schedule.push(item);persist();return copy(item);},()=>request('/schedule',{method:'POST',body})),
    removeSchedule:id=>choose(()=>{db.schedule=db.schedule.filter(x=>x.id!==id);persist();},()=>request('/schedule/'+enc(id),{method:'DELETE'})),
    students:id=>choose(()=>copy(db.students[id]||[]),()=>request('/classes/'+enc(id)+'/students')),
    attendance:(classId,date)=>choose(()=>copy(db.attendance[classId+':'+date]||{classId,date,records:[]}),()=>request('/attendance?classId='+enc(classId)+'&date='+enc(date))),
    saveAttendance:body=>choose(()=>{db.attendance[body.classId+':'+body.date]=copy(body);persist();return copy(body);},()=>request('/attendance',{method:'PUT',body})),
    discipline:()=>choose(()=>copy(db.discipline),()=>request('/discipline')),
    addDiscipline:body=>choose(()=>{const item={...copy(body),id:crypto.randomUUID()};db.discipline.unshift(item);persist();return copy(item);},()=>request('/discipline',{method:'POST',body})),
    session:()=>choose(()=>({user:{id:'demo',name:'مدرس نمونه',role:'teacher'}}),()=>request('/auth/session')),
    login:async body=>{const data=await request('/auth/login',{method:'POST',body});if(data?.token)setToken(data.token);return data;},
    register:body=>request('/auth/register',{method:'POST',body}),
    logout:async()=>{try{await request('/auth/logout',{method:'POST',body:{}});}finally{setToken('');}}
  };
})();

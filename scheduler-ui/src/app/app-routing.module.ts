import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SchedulesComponent } from './components/schedules/schedules.component';
import { ScheduleComponent } from './components/schedule/schedule.component';

const routes: Routes = [
  {path: "", component: SchedulesComponent},
  {path: "schedule/:id", component: ScheduleComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }

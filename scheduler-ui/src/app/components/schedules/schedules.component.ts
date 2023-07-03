import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { Page, Schedule } from 'src/app/models/schedule.model';
import { SchedulesService } from 'src/app/services/schedules.service';

@Component({
  selector: 'app-schedules',
  templateUrl: './schedules.component.html',
  styleUrls: ['./schedules.component.css']
})
export class SchedulesComponent {
  schedules: Schedule[] = [];
  page: Page;
  paginationState = {
    length: 10,
    pageIndex: 1,
    pageSize: 10,
    previousPageIndex: 0,
  };

  constructor(private schedulesService: SchedulesService, private router: Router) { }

  ngOnInit(): void {
    this.loadSchedules();
  }

  loadSchedules() {
    this.schedulesService.getSchedules(this.paginationState.pageSize, this.paginationState.pageIndex).subscribe(
      data => {
        this.schedules = data.schedules;
        this.page = data.page;
        this.paginationState.length = this.page.totalElements;
      }
    );

    console.log(this.paginationState);
  }

  openSchedule(id: string): void {
    this.router.navigate([`/schedule/${id}`]);
  }

  paginationChanged(event: any) {
    this.paginationState.length = event.length;
    this.paginationState.pageIndex = event.pageIndex + 1;
    this.paginationState.pageSize = event.pageSize;
    this.loadSchedules();
  }

  trigger(id: string) {
    this.schedulesService.triggerSchedule(id).subscribe(
      data => {
        console.log(data);
      }
    )
  }
}

import { Component } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { History, Page, Schedule } from 'src/app/models/schedule.model';
import { ScheduleService } from 'src/app/services/schedule.service';

@Component({
  selector: 'app-schedule',
  templateUrl: './schedule.component.html',
  styleUrls: ['./schedule.component.css']
})
export class ScheduleComponent {
  id: string = '';
  schedule: Schedule = new Schedule;
  history: History[] = [];
  page: Page;
  paginationState = {
    length: 10,
    pageIndex: 1,
    pageSize: 10,
    previousPageIndex: 0,
  };

  constructor(private scheduleService: ScheduleService, private route: ActivatedRoute) { }

  ngOnInit(): void {
    this.route.params.subscribe(params => {
      this.id = params['id'];
    });

    this.loadSchedule();
  }

  loadSchedule() {
    this.scheduleService.loadSchedule(this.id).subscribe(
      data => {
        this.schedule = data;
      }
    );
    this.loadHistory();
  }

  loadHistory() {
    this.scheduleService.loadHistory(this.id, this.paginationState.pageSize, this.paginationState.pageIndex).subscribe(
      data => {
        this.history = data.history;
        this.page = data.page;
        this.paginationState.length = this.page.totalElements;
      }
    )
  }

  isAcceptable(start: Date, end: Date): boolean {
    let enddate = new Date(end);
    let startdate = new Date(start);
    if (((enddate.getTime() - startdate.getTime()) / 1000) < 5)
      return true;
    return false;
  }

  paginationChanged(event: any) {
    this.paginationState.length = event.length;
    this.paginationState.pageIndex = event.pageIndex + 1;
    this.paginationState.pageSize = event.pageSize;
    this.loadHistory();
  }

  localTime(date: Date): string {
    return new Date(date).toLocaleString();
  }
}

import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'scheduler-ui';
  repoUrl = 'https://github.com/PSKP-95/scheduler';

  constructor(private router: Router) { }

  ngOnInit(): void {}

  goToGithub(): void {
    window.open(this.repoUrl, "_blank");
  }

  goToHome(): void {
    this.router.navigate([`/`]);
  }
}

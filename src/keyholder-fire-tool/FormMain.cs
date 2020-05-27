using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Data;
using System.Drawing;
using System.Text;
using System.Windows.Forms;

namespace keyholder_fire_tool
{
    public partial class FormMain : Form
    {
        public FormMain()
        {
            InitializeComponent();
        }

        private void linkApply_LinkClicked(object sender, LinkLabelLinkClickedEventArgs e)
        {
            System.Diagnostics.Process.Start(this.linkApply.Text);
        }

        private void linkAdmin_LinkClicked(object sender, LinkLabelLinkClickedEventArgs e)
        {
            System.Diagnostics.Process.Start(this.linkAdmin.Text);
        }

        private void openCS_Click(object sender, EventArgs e)
        {
            System.Diagnostics.Process.Start("cs.exe", "112.126.86.188 8957");
        }

        private void openClient_Click(object sender, EventArgs e)
        {
            System.Diagnostics.Process.Start("key-client.exe");
        }
    }
}

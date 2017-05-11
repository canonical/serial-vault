

# Get the PostgreSQL snap
snap:
	snap download postgresql96
	unsquashfs postgresql96_*.snap
	mv squashfs-root/etc ../install/
	mv squashfs-root/sbin ../install/
	mv squashfs-root/usr ../install/
